package raiju

import (
	"context"
	"errors"
	"fmt"
	"math"
	"math/rand"
	"os"
	"sort"
	"time"

	"github.com/nyonson/raiju/lightning"
)

//go:generate moq -stub -skip-ensure -out raiju_mock_test.go . lightninger

type lightninger interface {
	AddInvoice(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error)
	DescribeGraph(ctx context.Context) (*lightning.Graph, error)
	ForwardingHistory(ctx context.Context, since time.Time) ([]lightning.Forward, error)
	GetInfo(ctx context.Context) (*lightning.Info, error)
	GetChannel(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error)
	ListChannels(ctx context.Context) (lightning.Channels, error)
	SendPayment(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.Satoshi) (lightning.Satoshi, error)
	SetFees(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error
	SubscribeChannelUpdates(ctx context.Context) (<-chan lightning.Channels, <-chan error, error)
}

// Raiju app.
type Raiju struct {
	l lightninger
}

// New instance of raiju.
func New(l lightninger) Raiju {
	return Raiju{
		l: l,
	}
}

// LiquidityFees for channels with high, standard, and low liquidity.
type LiquidityFees struct {
	standard float64
}

// High liquidity channel fee to encourage more payments.
func (l LiquidityFees) High() lightning.FeePPM {
	return lightning.FeePPM(l.standard / 10)
}

// Standard liquidity channel fee.
func (l LiquidityFees) Standard() lightning.FeePPM {
	return lightning.FeePPM(l.standard)
}

// Low liquidity channel fee to discourage payments.
func (l LiquidityFees) Low() lightning.FeePPM {
	return lightning.FeePPM(l.standard * 10)
}

// NewLiquidityFees based on the standard fee.
func NewLiquidityFees(standard float64) LiquidityFees {
	return LiquidityFees{
		standard: standard,
	}
}

// RelativeNode has information on a node's graph characteristics relative to other nodes.
type RelativeNode struct {
	lightning.Node
	distance        int64
	distantNeigbors int64
	channels        int64
	capacity        lightning.Satoshi
	neighbors       []lightning.PubKey
}

// sortDistance sorts nodes by distance, distant neighbors, capacity, and channels
type sortDistance []RelativeNode

// Less is true if i is closer than j.
func (s sortDistance) Less(i, j int) bool {
	if s[i].distance != s[j].distance {
		return s[i].distance < s[j].distance
	}

	if s[i].distantNeigbors != s[j].distantNeigbors {
		return s[i].distantNeigbors < s[j].distantNeigbors
	}

	if s[i].capacity != s[j].capacity {
		return s[i].capacity < s[j].capacity
	}

	return s[i].channels < s[j].channels
}

// Swap nodes in slice.
func (s sortDistance) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

// Len of the relative node slice.
func (s sortDistance) Len() int {
	return len(s)
}

// CandidatesRequest contains necessary info to perform sorting across the network
type CandidatesRequest struct {
	// PubKey is the key of the root node to perform crawl from
	PubKey lightning.PubKey
	// MinCapcity filters nodes with a minimum satoshi capacity (sum of channels)
	MinCapacity lightning.Satoshi
	// MinChannels filters nodes with a minimum number of channels
	MinChannels int64
	// MinDistance filters nodes with a minimum distance (number of hops) from the root node
	MinDistance int64
	// MinDistantNeighbors filters nodes with a minimum number of distant neighbors
	MinDistantNeighbors int64
	// MinUpdated filters nodes which have not been updated since time
	MinUpdated time.Time
	// Assume channels to these pubkeys
	Assume []lightning.PubKey
	// Number of results
	Limit int64
	// Filter tor nodes
	Clearnet bool
}

// Candidates walks the lightning network from a specific node keeping track of distance (hops).
func (r Raiju) Candidates(ctx context.Context, request CandidatesRequest) ([]RelativeNode, error) {
	// default root node to local if no key supplied
	if request.PubKey == "" {
		info, err := r.l.GetInfo(ctx)
		if err != nil {
			return nil, err
		}

		request.PubKey = info.PubKey
	}

	// pull entire network graph from lnd
	channelGraph, err := r.l.DescribeGraph(ctx)

	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "network contains %d nodes total\n", len(channelGraph.Nodes))
	fmt.Fprintf(os.Stderr, "filtering candidates by capacity: %d, channels: %d, distance: %d, distant neighbors: %d\n", request.MinCapacity, request.MinChannels, request.MinDistance, request.MinDistantNeighbors)

	// initialize nodes map with static info
	nodes := make(map[lightning.PubKey]*RelativeNode, len(channelGraph.Nodes))

	for _, n := range channelGraph.Nodes {
		nodes[n.PubKey] = &RelativeNode{
			Node: n,
		}
	}

	// calculate node properties based on channels: neighbors, capacity, channels
	for _, e := range channelGraph.Edges {
		if nodes[e.Node1].neighbors != nil {
			nodes[e.Node1].neighbors = append(nodes[e.Node1].neighbors, e.Node2)
		} else {
			nodes[e.Node1].neighbors = []lightning.PubKey{e.Node2}
		}

		if nodes[e.Node2].neighbors != nil {
			nodes[e.Node2].neighbors = append(nodes[e.Node2].neighbors, e.Node1)
		} else {
			nodes[e.Node2].neighbors = []lightning.PubKey{e.Node1}
		}

		nodes[e.Node1].capacity += e.Capacity
		nodes[e.Node2].capacity += e.Capacity

		nodes[e.Node1].channels++
		nodes[e.Node2].channels++
	}

	// Add assumes to root node
	for _, c := range request.Assume {
		if _, ok := nodes[c]; !ok {
			return []RelativeNode{}, errors.New("candidate node does not exist")
		}

		if nodes[request.PubKey].neighbors != nil {
			nodes[request.PubKey].neighbors = append(nodes[request.PubKey].neighbors, c)
		} else {
			nodes[request.PubKey].neighbors = []lightning.PubKey{c}
		}

		if nodes[c].neighbors != nil {
			nodes[c].neighbors = append(nodes[c].neighbors, request.PubKey)
		} else {
			nodes[c].neighbors = []lightning.PubKey{request.PubKey}
		}
	}

	// BFS node graph to calculate distance from root node
	var distance int64 = 1
	visited := make(map[lightning.PubKey]bool)

	// handle strange case where root node doesn't exist for some reason...
	neighbors := make([]lightning.PubKey, 0)
	// initialize search from root node's neighbors
	if n, ok := nodes[request.PubKey]; ok {
		// root node has no distance to self
		n.distance = 0
		// mark root as visited
		visited[n.PubKey] = true
		neighbors = n.neighbors
	}

	for len(neighbors) > 0 {
		next := make([]lightning.PubKey, 0)
		for _, n := range neighbors {
			if !visited[n] {
				nodes[n].distance = distance
				visited[n] = true

				for _, neighbor := range nodes[n].neighbors {
					if !visited[neighbor] {
						next = append(next, neighbor)
					}
				}
			}
		}
		distance++
		neighbors = next
	}

	unfilteredSpan := make([]RelativeNode, 0)
	for _, v := range nodes {
		unfilteredSpan = append(unfilteredSpan, *v)
	}

	// hardcode what distance is considered "distant" for a neighbor
	const distantNeighborLimit int64 = 2

	// calculate number of distant neighbors per node
	for node := range unfilteredSpan {
		var count int64
		for _, neighbor := range unfilteredSpan[node].neighbors {
			if nodes[neighbor].distance > distantNeighborLimit {
				count++
			}
		}

		unfilteredSpan[node].distantNeigbors = count
	}

	// filter nodes by request conditions
	allCandidates := make([]RelativeNode, 0)
	for _, v := range unfilteredSpan {
		if v.capacity >= request.MinCapacity &&
			v.channels >= request.MinChannels &&
			v.distance >= request.MinDistance &&
			v.distantNeigbors >= request.MinDistantNeighbors &&
			v.Updated.After(request.MinUpdated) {
			if request.Clearnet {
				if v.Clearnet() {
					allCandidates = append(allCandidates, v)
				}
			} else {
				allCandidates = append(allCandidates, v)
			}
		}
	}

	sort.Sort(sort.Reverse(sortDistance(allCandidates)))

	if int64(len(allCandidates)) < request.Limit {
		return allCandidates, nil
	}

	candidates := allCandidates[:request.Limit]

	printNodes(candidates)

	return candidates, nil
}

// Fees to encourage a balanced channel.
//
// Daemon mode continuously updates policies as channel liquidity changes.
func (r Raiju) Fees(ctx context.Context, fees LiquidityFees, daemon bool) (map[lightning.ChannelID]lightning.ChannelLiquidityLevel, error) {
	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return map[lightning.ChannelID]lightning.ChannelLiquidityLevel{}, err
	}

	c, err := r.setFees(ctx, fees, channels)
	if err != nil {
		return map[lightning.ChannelID]lightning.ChannelLiquidityLevel{}, err
	}

	if daemon {
		cc, ce, err := r.l.SubscribeChannelUpdates(ctx)
		if err != nil {
			return map[lightning.ChannelID]lightning.ChannelLiquidityLevel{}, err
		}

		for {
			select {
			case channels = <-cc:
				_, err = r.setFees(ctx, fees, channels)
				if err != nil {
					fmt.Fprintf(os.Stderr, "error setting fees %v\n", err)
				}
			case err := <-ce:
				fmt.Fprintf(os.Stderr, "error listening to channel updates %v\n", err)
			}
		}
	}

	return c, nil
}

// setFees on channels who's liquidity has changed, return updated channels and their new liquidity level.
func (r Raiju) setFees(ctx context.Context, fees LiquidityFees, channels lightning.Channels) (map[lightning.ChannelID]lightning.ChannelLiquidityLevel, error) {
	updates := make(map[lightning.ChannelID]lightning.ChannelLiquidityLevel)
	// update channel fees based on liquidity, but only change if necessary
	for _, c := range channels {
		switch c.LiquidityLevel() {
		case lightning.LowLiquidity:
			if c.LocalFee != fees.Low() {
				fmt.Fprintf(os.Stderr, "channel %d now has low liquidity %f, setting fee to %f\n", c.ChannelID, c.Liquidity(), fees.Low())
				err := r.l.SetFees(ctx, c.ChannelID, fees.Low())
				if err != nil {
					fmt.Fprintf(os.Stderr, "error updating fees %v\n", err)
				} else {
					updates[c.ChannelID] = lightning.LowLiquidity
				}
			}
		case lightning.StandardLiquidity:
			if c.LocalFee != fees.Standard() {
				fmt.Fprintf(os.Stderr, "channel %d now has standard liquidity %f, setting fee to %f\n", c.ChannelID, c.Liquidity(), fees.Standard())
				err := r.l.SetFees(ctx, c.ChannelID, fees.Standard())
				if err != nil {
					fmt.Fprintf(os.Stderr, "error updating fees %v\n", err)
				} else {
					updates[c.ChannelID] = lightning.StandardLiquidity
				}
			}
		case lightning.HighLiquidity:
			if c.LocalFee != fees.High() {
				fmt.Fprintf(os.Stderr, "channel %d now has high liquidity %f, setting fee to %f\n", c.ChannelID, c.Liquidity(), fees.High())
				err := r.l.SetFees(ctx, c.ChannelID, fees.High())
				if err != nil {
					fmt.Fprintf(os.Stderr, "error updating fees %v\n", err)
				} else {
					updates[c.ChannelID] = lightning.HighLiquidity
				}
			}
		}
	}

	return updates, nil
}

// Rebalance liquidity out of outChannelID and in through lastHopPubkey and returns the percent of capacity rebalanced.
//
// The amount of sats rebalanced is based on the capacity of the out channel. Each rebalance attempt will try to move
// stepPercent worth of sats. A maximum of maxPercent of sats will be moved. The maxFee in ppm controls the amount
// willing to pay for rebalance.
func (r Raiju) Rebalance(ctx context.Context, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, stepPercent float64, maxPercent float64, maxFee lightning.FeePPM) (float64, lightning.Satoshi, error) {
	// calculate invoice value
	c, err := r.l.GetChannel(ctx, outChannelID)
	if err != nil {
		return 0, 0, err
	}

	amount := int64(float64(c.Capacity) * stepPercent / 100)
	// add 0.5 so rounds up to at least 1
	maxFeeSats := int64(math.Round(maxFee.Rate()*float64(amount) + 0.5))

	var percentRebalanced float64
	var totalFeePaid lightning.Satoshi

	for percentRebalanced < maxPercent {
		fmt.Fprintf(os.Stderr, "attempting rebalance %d sats out of %d to %s with a %d sats max fee...\n", amount, outChannelID, lastHopPubKey, maxFeeSats)
		// create and pay invoice
		invoice, err := r.l.AddInvoice(ctx, lightning.Satoshi(amount))
		if err != nil {
			fmt.Fprintf(os.Stderr, "rebalance failed\n")
			return percentRebalanced, totalFeePaid, nil
		}
		feePaid, err := r.l.SendPayment(ctx, invoice, outChannelID, lastHopPubKey, lightning.Satoshi(maxFeeSats))
		if err != nil {
			fmt.Fprintf(os.Stderr, "rebalance failed\n")
			return percentRebalanced, totalFeePaid, nil
		}
		fmt.Fprintf(os.Stderr, "rebalance success %d sats out of %d to %s for a %d sats fee\n", amount, outChannelID, lastHopPubKey, feePaid)
		percentRebalanced += stepPercent
		totalFeePaid += feePaid
		fmt.Fprintf(os.Stderr, "rebalance has moved %f percent of max %f percent of the channel capacity\n", percentRebalanced, maxPercent)
	}

	return percentRebalanced, totalFeePaid, nil
}

// RebalanceAll channels.
func (r Raiju) RebalanceAll(ctx context.Context, stepPercent float64, maxPercent float64, maxFee lightning.FeePPM) error {
	local, err := r.l.GetInfo(ctx)
	if err != nil {
		return err
	}

	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return err
	}

	hlcs := channels.HighLiquidity()
	llcs := channels.LowLiquidity()

	// Shuffle arrays so different combos are tried
	rand.Shuffle(len(hlcs), func(i, j int) {
		hlcs[i], hlcs[j] = hlcs[j], hlcs[i]
	})
	rand.Shuffle(len(llcs), func(i, j int) {
		llcs[i], llcs[j] = llcs[j], llcs[i]
	})

	var totalFeePaid lightning.Satoshi

	// Roll through high liquidity channels and try to push things through the low liquidity ones.
	for _, h := range hlcs {
		fmt.Fprintf(os.Stderr, "channel %d with %s has high liquidity, attempting to rebalancing into low liquidity channels\n", h.ChannelID, h.RemoteNode.Alias)
		percentRebalanced := float64(0)
		for _, l := range llcs {
			// get the non-local node of the channel
			lastHopPubkey := l.Node1
			if lastHopPubkey == local.PubKey {
				lastHopPubkey = l.Node2
			}

			// Check that the channel is still low on liquidity since rebalancing started
			ul, err := r.l.GetChannel(ctx, l.ChannelID)
			if err != nil {
				return err
			}
			if ul.LiquidityLevel() == lightning.LowLiquidity {
				// don't really care if error or not, just continue on
				p, f, _ := r.Rebalance(ctx, h.ChannelID, lastHopPubkey, stepPercent, (maxPercent - percentRebalanced), maxFee)
				percentRebalanced += p
				totalFeePaid += f
			} else {
				fmt.Fprintf(os.Stderr, "channel %d with %s no longer has low liquidity\n", ul.ChannelID, ul.RemoteNode.Alias)
			}
		}
		fmt.Fprintf(os.Stderr, "rebalanced %f percent of channel %d with %s\n", percentRebalanced, h.ChannelID, h.RemoteNode.Alias)
	}

	fmt.Fprintf(os.Stderr, "rebalanced channels paying a total fee of %d sats\n", totalFeePaid)

	return nil
}

// Reaper calculates inefficient channels which should be closed.
func (r Raiju) Reaper(ctx context.Context) (lightning.Channels, error) {
	// pull the last month of forwards
	forwards, err := r.l.ForwardingHistory(ctx, time.Now().AddDate(0, -1, 0))
	if err != nil {
		return lightning.Channels{}, err
	}

	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return lightning.Channels{}, err
	}

	// initialize tracker
	m := make(map[lightning.ChannelID]bool)
	for _, c := range channels {
		m[c.ChannelID] = false
	}

	for _, f := range forwards {
		m[f.ChannelIn] = true
		m[f.ChannelOut] = true
	}

	inefficient := make(lightning.Channels, 0)
	for _, c := range channels {
		if !m[c.ChannelID] {
			inefficient = append(inefficient, c)
		}
	}

	printChannels(inefficient)

	return inefficient, nil
}
