package raiju

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"sort"
	"time"

	"github.com/nyonson/raiju/lightning"
)

//go:generate gotests -w -exported raiju.go
//go:generate moq -stub -skip-ensure -out raiju_mock_test.go . lightninger

type lightninger interface {
	AddInvoice(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error)
	DescribeGraph(ctx context.Context) (*lightning.Graph, error)
	ForwardingHistory(ctx context.Context, since time.Time) ([]lightning.Forward, error)
	GetInfo(ctx context.Context) (*lightning.Info, error)
	GetChannel(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error)
	ListChannels(ctx context.Context) (lightning.Channels, error)
	SendPayment(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.FeePPM) (lightning.Satoshi, error)
	SetFees(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error
	SubscribeChannelUpdates(ctx context.Context) (<-chan lightning.Channels, <-chan error, error)
}

// Raiju app.
type Raiju struct {
	l lightninger
	f LiquidityFees
}

// New instance of raiju.
func New(l lightninger, r LiquidityFees) Raiju {
	return Raiju{
		l: l,
		f: r,
	}
}

// RelativeNode has information on a node's graph characteristics relative to other nodes.
type RelativeNode struct {
	lightning.Node
	Distance        int64
	DistantNeigbors int64
	Channels        int64
	Capacity        lightning.Satoshi
	Neighbors       []lightning.PubKey
}

// sortDistance sorts nodes by distance, distant neighbors, capacity, and channels
type sortDistance []RelativeNode

// Less is true if i is closer than j.
func (s sortDistance) Less(i, j int) bool {
	if s[i].Distance != s[j].Distance {
		return s[i].Distance < s[j].Distance
	}

	if s[i].DistantNeigbors != s[j].DistantNeigbors {
		return s[i].DistantNeigbors < s[j].DistantNeigbors
	}

	if s[i].Capacity != s[j].Capacity {
		return s[i].Capacity < s[j].Capacity
	}

	return s[i].Channels < s[j].Channels
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

	// initialize nodes map with static info
	nodes := make(map[lightning.PubKey]*RelativeNode, len(channelGraph.Nodes))

	for _, n := range channelGraph.Nodes {
		nodes[n.PubKey] = &RelativeNode{
			Node: n,
		}
	}

	// calculate node properties based on channels: neighbors, capacity, channels
	for _, e := range channelGraph.Edges {
		if nodes[e.Node1].Neighbors != nil {
			nodes[e.Node1].Neighbors = append(nodes[e.Node1].Neighbors, e.Node2)
		} else {
			nodes[e.Node1].Neighbors = []lightning.PubKey{e.Node2}
		}

		if nodes[e.Node2].Neighbors != nil {
			nodes[e.Node2].Neighbors = append(nodes[e.Node2].Neighbors, e.Node1)
		} else {
			nodes[e.Node2].Neighbors = []lightning.PubKey{e.Node1}
		}

		nodes[e.Node1].Capacity += e.Capacity
		nodes[e.Node2].Capacity += e.Capacity

		nodes[e.Node1].Channels++
		nodes[e.Node2].Channels++
	}

	// Add assumes to root node
	for _, c := range request.Assume {
		if _, ok := nodes[c]; !ok {
			return []RelativeNode{}, errors.New("candidate node does not exist")
		}

		if nodes[request.PubKey].Neighbors != nil {
			nodes[request.PubKey].Neighbors = append(nodes[request.PubKey].Neighbors, c)
		} else {
			nodes[request.PubKey].Neighbors = []lightning.PubKey{c}
		}

		if nodes[c].Neighbors != nil {
			nodes[c].Neighbors = append(nodes[c].Neighbors, request.PubKey)
		} else {
			nodes[c].Neighbors = []lightning.PubKey{request.PubKey}
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
		n.Distance = 0
		// mark root as visited
		visited[n.PubKey] = true
		neighbors = n.Neighbors
	}

	for len(neighbors) > 0 {
		next := make([]lightning.PubKey, 0)
		for _, n := range neighbors {
			if !visited[n] {
				nodes[n].Distance = distance
				visited[n] = true

				for _, neighbor := range nodes[n].Neighbors {
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
		for _, neighbor := range unfilteredSpan[node].Neighbors {
			if nodes[neighbor].Distance > distantNeighborLimit {
				count++
			}
		}

		unfilteredSpan[node].DistantNeigbors = count
	}

	// filter nodes by request conditions
	allCandidates := make([]RelativeNode, 0)
	for _, v := range unfilteredSpan {
		if v.Capacity >= request.MinCapacity &&
			v.Channels >= request.MinChannels &&
			v.Distance >= request.MinDistance &&
			v.DistantNeigbors >= request.MinDistantNeighbors &&
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

	candidates := allCandidates
	if int64(len(allCandidates)) >= request.Limit {
		candidates = allCandidates[:request.Limit]
	}

	return candidates, nil
}

// Fees to encourage a balanced channel.
//
// Fees are initially set across all channels and then continuously updated as channel liquidity changes.
func (r Raiju) Fees(ctx context.Context) (chan map[lightning.ChannelID]lightning.FeePPM, chan error, error) {
	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return nil, nil, err
	}

	// buffer the channel for the first update
	updates := make(chan map[lightning.ChannelID]lightning.FeePPM, 1)
	errors := make(chan error)
	// make sure updated at least once
	u, err := r.setFees(ctx, channels)
	if err != nil {
		return nil, nil, err
	}

	updates <- u

	// listen for channel updates to keep fees in sync
	cc, ce, err := r.l.SubscribeChannelUpdates(ctx)
	if err != nil {
		return nil, nil, err
	}

	go func() {
		for {
			select {
			case channels = <-cc:
				u, err = r.setFees(ctx, channels)
				if err != nil {
					errors <- fmt.Errorf("error setting fees: %w", err)
				} else {
					updates <- u
				}
			case err := <-ce:
				errors <- fmt.Errorf("error listening to channel updates: %w", err)
			}
		}
	}()

	return updates, errors, nil
}

// setFees on channels who's liquidity has changed, return updated channels and their new liquidity level.
func (r Raiju) setFees(ctx context.Context, channels lightning.Channels) (map[lightning.ChannelID]lightning.FeePPM, error) {
	updates := map[lightning.ChannelID]lightning.FeePPM{}
	// update channel fees based on liquidity, but only change if necessary
	for _, c := range channels {
		fee := r.f.Fee(c)
		if c.LocalFee != fee && !c.Private {
			err := r.l.SetFees(ctx, c.ChannelID, fee)
			if err != nil {
				return map[lightning.ChannelID]lightning.FeePPM{}, err
			}
			updates[c.ChannelID] = fee
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

	var percentRebalanced float64
	var totalFeePaid lightning.Satoshi

	for percentRebalanced < maxPercent {
		// create and pay invoice
		invoice, err := r.l.AddInvoice(ctx, lightning.Satoshi(amount))
		if err != nil {
			return 0, 0, fmt.Errorf("error creating circular rebalance invoice: %w", err)
		}
		feePaid, err := r.l.SendPayment(ctx, invoice, outChannelID, lastHopPubKey, maxFee)
		// not expecting rebalance payments to work all that often, so just short circuit and return what has been done
		if err != nil {
			return percentRebalanced, totalFeePaid, nil
		}
		percentRebalanced += stepPercent
		totalFeePaid += feePaid
	}

	return percentRebalanced, totalFeePaid, nil
}

// RebalanceAll high local liquidity channels into low liquidity channels, return percent rebalanced per channel attempted.
func (r Raiju) RebalanceAll(ctx context.Context, stepPercent float64, maxPercent float64) (map[lightning.ChannelID]float64, error) {
	local, err := r.l.GetInfo(ctx)
	if err != nil {
		return map[lightning.ChannelID]float64{}, err
	}

	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return map[lightning.ChannelID]float64{}, err
	}

	hlcs, llcs := r.f.RebalanceChannels(channels)

	// Shuffle arrays so different combos are tried
	rand.Shuffle(len(hlcs), func(i, j int) {
		hlcs[i], hlcs[j] = hlcs[j], hlcs[i]
	})

	var totalFeePaid lightning.Satoshi
	rebalanced := map[lightning.ChannelID]float64{}

	// Roll through high liquidity channels and try to push things through the low liquidity ones.
	for _, h := range hlcs {
		percentRebalanced := float64(0)

		// reshuffle low liquidity channels each time
		rand.Shuffle(len(llcs), func(i, j int) {
			llcs[i], llcs[j] = llcs[j], llcs[i]
		})
		for _, l := range llcs {
			// get the non-local node of the channel
			lastHopPubkey := l.Node1
			if lastHopPubkey == local.PubKey {
				lastHopPubkey = l.Node2
			}

			// Check that the channel would still be low liquidity to avoid the risk of paying a high fee
			// to rebalance and then a standard payment cancels out the liquidity
			ul, err := r.l.GetChannel(ctx, l.ChannelID)
			if err != nil {
				return map[lightning.ChannelID]float64{}, err
			}
			potentialLocal := lightning.Satoshi(float64(h.Capacity) * maxPercent)
			if r.f.PotentialFee(ul, potentialLocal) != r.f.Fee(ul) {
				p, f, err := r.Rebalance(ctx, h.ChannelID, lastHopPubkey, stepPercent, (maxPercent - percentRebalanced), r.f.RebalanceFee())
				if err != nil {
					return map[lightning.ChannelID]float64{}, err
				}

				percentRebalanced += p
				totalFeePaid += f
			}
		}
		rebalanced[h.ChannelID] = percentRebalanced
	}

	return rebalanced, nil
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

	return inefficient, nil
}
