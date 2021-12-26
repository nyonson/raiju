package raiju

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/btcsuite/btcutil"

	"git.sr.ht/~yonson/raiju/lightning"
)

type client interface {
	GetInfo(ctx context.Context) (*lightning.Info, error)
	DescribeGraph(ctx context.Context) (*lightning.Graph, error)
}

type Raiju struct {
	client client
}

func New(client client) Raiju {
	return Raiju{
		client: client,
	}
}

// BtcToSat returns the btc amount in satoshis
func BtcToSat(btc float64) int {
	return int(btc * 100000000)
}

// RelativeNode has information on a node's graph charactericts relative to other nodes.
type RelativeNode struct {
	lightning.Node
	distance        int
	distantNeigbors int
	channels        int
	capacity        float64
	neighbors       []string
}

// sortDistance sorts nodes by distance, distant neigbors, capacity, and channels
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
	MinCapacity float64
	// MinChannels filters nodes with a minimum number of channels
	MinChannels int
	// MinDistance filters nodes with a minumum distance (number of hops) from the root node
	MinDistance int
	// MinNeighborDistance is the distance required for a node to be considered a distanct neighbor
	MinNeighborDistance int
	// MinUpdated filters nodes which have not been updated since time
	MinUpdated time.Time
	// Assume channels to these pubkeys
	Assume []string
	// Number of results
	Limit int
}

// Candidates walks the lightning network from a specific node keeping track of distance (hops).
func (r Raiju) Candidates(ctx context.Context, request CandidatesRequest) ([]RelativeNode, error) {
	// default root node to local if no key supplied
	if request.Pubkey == "" {
		info, err := r.client.GetInfo(ctx)
		if err != nil {
			return nil, err
		}

		request.Pubkey = info.Pubkey
	}

	// pull entire network graph from lnd
	channelGraph, err := r.client.DescribeGraph(ctx)

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

		nodes[e.Node1].capacity += e.Capacity.ToUnit(btcutil.AmountSatoshi)
		nodes[e.Node2].capacity += e.Capacity.ToUnit(btcutil.AmountSatoshi)

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
	count := 1
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
		count := 0
		for _, n := range unfilteredSpan[i].neighbors {
			if nodes[n].distance > request.MinNeighborDistance {
				count++
			}
		}

		unfilteredSpan[i].distantNeigbors = count
	}

	// filter nodes by request minimums
	span := make([]RelativeNode, 0)
	for _, v := range unfilteredSpan {
		if v.capacity >= request.MinCapacity && v.channels >= request.MinChannels && v.distance >= request.MinDistance && v.Updated.After(request.MinUpdated) {
			span = append(span, v)
		}
	}

	sort.Sort(sort.Reverse(sortDistance(span)))

	if len(span) < request.Limit {
		return span, nil
	}

	return span[:request.Limit], nil
}
