package raiju

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/nyonson/raiju/lightning"
)

type lightninger interface {
	AddInvoice(ctx context.Context, amount int64) (string, error)
	DescribeGraph(ctx context.Context) (*lightning.Graph, error)
	GetInfo(ctx context.Context) (*lightning.Info, error)
	GetChannel(ctx context.Context, channelID uint64) (lightning.Channel, error)
	ListChannels(ctx context.Context) ([]lightning.Channel, error)
	SendPayment(ctx context.Context, invoice string, outChannelID uint64, lastHopPubkey string, maxFee int64) error
	SetFees(ctx context.Context, channelID uint64, fee int64) error
}

type Raiju struct {
	l lightninger
}

func New(l lightninger) Raiju {
	return Raiju{
		l: l,
	}
}

// BtcToSat returns the bitcoin amount in satoshis
func BtcToSat(btc float64) int64 {
	return int64(btc * 100000000)
}

// SatsToBtc returns the satoshis amount in bitcoin
func SatsToBtc(sats int64) float64 {
	return float64(sats) / 100000000
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
func (r Raiju) Fees(ctx context.Context, standardFee int64) error {
	channels, err := r.l.ListChannels(ctx)
	if err != nil {
		return err
	}

	lowLiqudityChannels, standardLiquidityChannels, highLiquidityChannels := lightning.ChannelLiquidities(channels)

	// encourage relatively more inbound txs by raising local fees.
	lowLiquidityFee := standardFee * 10
	// encourage relatively more outbound txs by lowering local fees.
	highLiquidityFee := standardFee / 10

	for _, c := range lowLiqudityChannels {
		liquidity := c.Local.ToUnit(btcutil.AmountSatoshi) / (c.Local.ToUnit(btcutil.AmountSatoshi) + c.Remote.ToUnit(btcutil.AmountSatoshi)) * 100
		fmt.Fprintf(os.Stderr, "channel %d has low liquidity %f setting fee to %d\n", c.ChannelID, liquidity, lowLiquidityFee)
		r.l.SetFees(ctx, c.ChannelID, lowLiquidityFee)
	}

	for _, c := range standardLiquidityChannels {
		liquidity := c.Local.ToUnit(btcutil.AmountSatoshi) / (c.Local.ToUnit(btcutil.AmountSatoshi) + c.Remote.ToUnit(btcutil.AmountSatoshi)) * 100
		fmt.Fprintf(os.Stderr, "channel %d has standard liquidity %f setting fee to %d\n", c.ChannelID, liquidity, standardFee)
		r.l.SetFees(ctx, c.ChannelID, standardFee)
	}

	for _, c := range highLiquidityChannels {
		liquidity := c.Local.ToUnit(btcutil.AmountSatoshi) / (c.Local.ToUnit(btcutil.AmountSatoshi) + c.Remote.ToUnit(btcutil.AmountSatoshi)) * 100
		fmt.Fprintf(os.Stderr, "channel %d has high liquidity %f setting fee to %d\n", c.ChannelID, liquidity, highLiquidityFee)
		r.l.SetFees(ctx, c.ChannelID, highLiquidityFee)
	}

	return nil
}

// https://github.com/lightning/bolts/blob/master/04-onion-routing.md#non-strict-forwarding
func (r Raiju) Rebalance(ctx context.Context, outChannelID uint64, lastHopPubkey string, percent int64, maxFee int64) error {
	// calculate invoice value
	c, err := r.l.GetChannel(ctx, outChannelID)
	if err != nil {
		return err
	}

	amount := int64(c.Capacity.ToUnit(btcutil.AmountSatoshi) * (float64(percent) / 100))

	// create invoice
	invoice, err := r.l.AddInvoice(ctx, amount)
	if err != nil {
		return err
	}

	// pay invoice
	return r.l.SendPayment(ctx, invoice, outChannelID, lastHopPubkey, maxFee)
}

// RebalanceAllRequest contains necessary info to perform circular rebalance
type RebalanceAllRequest struct {
	// MaxFee in sats to pay for a rebalance
	MaxFee int64
}

func (r Raiju) RebalanceAll(ctx context.Context, request RebalanceAllRequest) error {
	// channels, err := r.l.ListChannels(ctx)
	// if err != nil {
	// 	return err
	// }
	//
	// lowLiqudityChannels, _, highLiquidityChannels := lightning.ChannelLiquidities(channels)
	//
	// for _, l := range lowLiqudityChannels {
	// }

	return nil
}
