package raiju

import (
	"context"
	"fmt"
	"math"
	"os"
	"sort"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/nyonson/raiju/lightning"
)

type lightninger interface {
	AddInvoice(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error)
	DescribeGraph(ctx context.Context) (*lightning.Graph, error)
	ForwardingHistory(ctx context.Context, since time.Time) ([]lightning.Forward, error)
	GetInfo(ctx context.Context) (*lightning.Info, error)
	GetChannel(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error)
	ListChannels(ctx context.Context) (lightning.Channels, error)
	SendPayment(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubkey string, maxFee lightning.Satoshi) (lightning.Satoshi, error)
	SetFees(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error
}

type Raiju struct {
	l lightninger
}

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
	capacity        int64
	neighbors       []string
}

// sortDistance sorts nodes by distance, distant neighbors, capacity, and channels
type sortDistance []RelativeNode

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

func (s sortDistance) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortDistance) Len() int {
	return len(s)
}

// CandidatesRequest contains necessary info to perform sorting across the network
type CandidatesRequest struct {
	// Pubkey is the key of the root node to perform crawl from
	Pubkey string
	// MinCapcity filters nodes with a minimum satoshi capacity (sum of channels)
	MinCapacity int64
	// MinChannels filters nodes with a minimum number of channels
	MinChannels int64
	// MinDistance filters nodes with a minumum distance (number of hops) from the root node
	MinDistance int64
	// MinNeighborDistance is the distance required for a node to be considered a distanct neighbor
	MinNeighborDistance int64
	// MinUpdated filters nodes which have not been updated since time
	MinUpdated time.Time
	// Assume channels to these pubkeys
	Assume []string
	// Number of results
	Limit int64
	// Filter tor nodes
	Clearnet bool
}

// Candidates walks the lightning network from a specific node keeping track of distance (hops).
func (r Raiju) Candidates(ctx context.Context, request CandidatesRequest) ([]RelativeNode, error) {
	// default root node to local if no key supplied
	if request.Pubkey == "" {
		info, err := r.l.GetInfo(ctx)
		if err != nil {
			return nil, err
		}

		request.Pubkey = info.Pubkey
	}

	// pull entire network graph from lnd
	channelGraph, err := r.l.DescribeGraph(ctx)

	if err != nil {
		return nil, err
	}

	fmt.Fprintf(os.Stderr, "network contains %d nodes total\n", len(channelGraph.Nodes))

	// initialize nodes map with static info
	nodes := make(map[string]*RelativeNode, len(channelGraph.Nodes))

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
			nodes[e.Node1].neighbors = []string{e.Node2}
		}

		if nodes[e.Node2].neighbors != nil {
			nodes[e.Node2].neighbors = append(nodes[e.Node2].neighbors, e.Node1)
		} else {
			nodes[e.Node2].neighbors = []string{e.Node1}
		}

		nodes[e.Node1].capacity += int64(e.Capacity)
		nodes[e.Node2].capacity += int64(e.Capacity)

		nodes[e.Node1].channels++
		nodes[e.Node2].channels++
	}

	// Add assumes to root node
	for _, c := range request.Assume {
		if _, ok := nodes[c]; !ok {
			fmt.Fprintf(os.Stderr, "candidate node does not exist: %s\n", c)
			continue
		}

		if nodes[request.Pubkey].neighbors != nil {
			nodes[request.Pubkey].neighbors = append(nodes[request.Pubkey].neighbors, c)
		} else {
			nodes[request.Pubkey].neighbors = []string{c}
		}

		if nodes[c].neighbors != nil {
			nodes[c].neighbors = append(nodes[c].neighbors, request.Pubkey)
		} else {
			nodes[c].neighbors = []string{request.Pubkey}
		}
	}

	// BFS node graph to calculate distance from root node
	var count int64 = 1
	visited := make(map[string]bool)

	current := nodes[request.Pubkey].neighbors
	for len(current) > 0 {
		next := make([]string, 0)
		for _, n := range current {
			if !visited[n] {
				nodes[n].distance = count
				visited[n] = true

				for _, neighbor := range nodes[n].neighbors {
					if !visited[neighbor] {
						next = append(next, neighbor)
					}
				}
			}
		}
		count++
		current = next
	}

	unfilteredSpan := make([]RelativeNode, 0)
	for _, v := range nodes {
		unfilteredSpan = append(unfilteredSpan, *v)
	}

	// calculate number of distant neighbors per node
	for i := range unfilteredSpan {
		var count int64 = 0
		for _, n := range unfilteredSpan[i].neighbors {
			if nodes[n].distance > request.MinNeighborDistance {
				count++
			}
		}

		unfilteredSpan[i].distantNeigbors = count
	}

	// filter nodes by request conditions
	span := make([]RelativeNode, 0)
	for _, v := range unfilteredSpan {
		if v.capacity >= request.MinCapacity &&
			v.channels >= request.MinChannels &&
			v.distance >= request.MinDistance &&
			v.Updated.After(request.MinUpdated) {
			if request.Clearnet {
				if v.Clearnet() {
					span = append(span, v)
				}
			} else {
				span = append(span, v)
			}
		}
	}

	sort.Sort(sort.Reverse(sortDistance(span)))

	if int64(len(span)) < request.Limit {
		return span, nil
	}

	return span[:request.Limit], nil
}

// Fees to encourage a balanced channel.
func (r Raiju) Fees(ctx context.Context, fees LiquidityFees) error {
	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return err
	}

	for _, c := range channels {
		switch c.Liquidity() {
		case lightning.LowLiquidity:
			liquidity := c.LocalBalance.ToUnit(btcutil.AmountSatoshi) / (c.LocalBalance.ToUnit(btcutil.AmountSatoshi) + c.RemoteBalance.ToUnit(btcutil.AmountSatoshi)) * 100
			fmt.Fprintf(os.Stderr, "channel %d has low liquidity %f setting fee to %f\n", c.ChannelID, liquidity, fees.Low())
			r.l.SetFees(ctx, c.ChannelID, fees.Low())
		case lightning.StandardLiquidity:
			liquidity := c.LocalBalance.ToUnit(btcutil.AmountSatoshi) / (c.LocalBalance.ToUnit(btcutil.AmountSatoshi) + c.RemoteBalance.ToUnit(btcutil.AmountSatoshi)) * 100
			fmt.Fprintf(os.Stderr, "channel %d has standard liquidity %f setting fee to %f\n", c.ChannelID, liquidity, fees.Standard())
			r.l.SetFees(ctx, c.ChannelID, fees.Standard())
		case lightning.HighLiquidity:
			liquidity := c.LocalBalance.ToUnit(btcutil.AmountSatoshi) / (c.LocalBalance.ToUnit(btcutil.AmountSatoshi) + c.RemoteBalance.ToUnit(btcutil.AmountSatoshi)) * 100
			fmt.Fprintf(os.Stderr, "channel %d has high liquidity %f setting fee to %f\n", c.ChannelID, liquidity, fees.High())
			r.l.SetFees(ctx, c.ChannelID, fees.High())
		}

	}

	return nil
}

// Rebalance channel.
func (r Raiju) Rebalance(ctx context.Context, outChannelID lightning.ChannelID, lastHopPubkey string, percent float64, max lightning.FeePPM) error {
	// calculate invoice value
	c, err := r.l.GetChannel(ctx, outChannelID)
	if err != nil {
		return err
	}

	amount := int64(c.Capacity.ToUnit(btcutil.AmountSatoshi) * (percent / 100))
	// add 0.5 so rounds up to at least 1
	maxFee := int64(math.Round(max.Rate()*float64(amount) + 0.5))

	fmt.Fprintf(os.Stderr, "attempting rebalance %d sats out of %d to %s with a %d max fee...\n", amount, outChannelID, lastHopPubkey, maxFee)

	// create invoice
	invoice, err := r.l.AddInvoice(ctx, lightning.Satoshi(amount))
	if err != nil {
		return err
	}

	// pay invoice
	fee, err := r.l.SendPayment(ctx, invoice, outChannelID, lastHopPubkey, lightning.Satoshi(maxFee))
	if err != nil {
		fmt.Fprintf(os.Stderr, "rebalance failed\n")
		return err
	}

	fmt.Fprintf(os.Stderr, "rebalance success %d sats out of %d to %s for a %d sats fee\n", amount, outChannelID, lastHopPubkey, fee)

	return nil
}

// RebalanceAll channels.
func (r Raiju) RebalanceAll(ctx context.Context, percent float64, max lightning.FeePPM) error {
	local, err := r.l.GetInfo(ctx)
	if err != nil {
		return err
	}

	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return err
	}

	// Roll through high liquidity channels and try to push things through the low liquidity ones.
	for _, h := range channels.HighLiquidity() {
		for _, l := range channels.LowLiquidity() {
			// get the non-local node of the channel
			lastHopPubkey := l.Node1
			if lastHopPubkey == local.Pubkey {
				lastHopPubkey = l.Node2
			}

			// Check that the channel is still low on liquidity since rebalancing started
			ul, err := r.l.GetChannel(ctx, l.ChannelID)
			if err != nil {
				return err
			}
			if ul.Liquidity() == lightning.LowLiquidity {
				// don't really care if error or not, just continue on
				r.Rebalance(ctx, h.ChannelID, lastHopPubkey, percent, max)
			}
		}
	}

	return nil
}

// Reaper calculates inefficient channels which should be closed.
func (r Raiju) Reaper(ctx context.Context) error {
	// pull the last month of forwards
	forwards, err := r.l.ForwardingHistory(ctx, time.Now().AddDate(0, -1, 0))
	if err != nil {
		return err
	}

	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return err
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

	PrintChannels(inefficient)

	return nil
}
