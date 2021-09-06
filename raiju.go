package raiju

import (
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/btcsuite/btcutil"
	"github.com/lightninglabs/lndclient"
	"github.com/rodaine/table"
)

type Client interface {
	GetInfo(ctx context.Context) (*lndclient.Info, error)
	DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error)
}

// App context
type App struct {
	Client  Client
	Log     *log.Logger
	Verbose bool
}

// BtcToSat returns the btc amount in satoshis
func BtcToSat(btc float64) int {
	return int(btc * 100000000)
}

// PrintBtcToSat prints the btc amount in satoshis to stdout
func PrintBtcToSat(btc float64) {
	fmt.Fprintln(os.Stdout, BtcToSat(btc))
}

// Node in the lightning graph network with computed properties
type Node struct {
	pubkey          string
	alias           string
	distance        int
	distantNeigbors int
	channels        int
	capacity        float64
	updated         time.Time
	neighbors       []string
}

// sortDistance sorts nodes by distance, distant neigbors, capacity, and channels
type sortDistance []Node

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

// Candidates walks the lightning network from a specific node keeping track of distance (hops)
func Candidates(app App, request CandidatesRequest) ([]Node, error) {
	// default root node to local lnd if no key supplied
	if request.Pubkey == "" {
		pk, err := localPubkey(app)
		if err != nil {
			return nil, err
		}

		request.Pubkey = pk
	}

	// pull entire network graph from lnd
	channelGraph, err := app.Client.DescribeGraph(context.Background(), false)

	if err != nil {
		return nil, err
	}

	if app.Verbose {
		app.Log.Printf("network contains %d nodes total\n", len(channelGraph.Nodes))
	}

	// initialize nodes map with static info
	nodes := make(map[string]*Node, len(channelGraph.Nodes))

	for _, n := range channelGraph.Nodes {
		nodes[n.PubKey.String()] = &Node{pubkey: n.PubKey.String(), alias: n.Alias, updated: n.LastUpdate}
	}

	// calculate node properties based on channels: neighbors, capacity, channels
	for _, e := range channelGraph.Edges {
		if nodes[e.Node1.String()].neighbors != nil {
			nodes[e.Node1.String()].neighbors = append(nodes[e.Node1.String()].neighbors, e.Node2.String())
		} else {
			nodes[e.Node1.String()].neighbors = []string{e.Node2.String()}
		}

		if nodes[e.Node2.String()].neighbors != nil {
			nodes[e.Node2.String()].neighbors = append(nodes[e.Node2.String()].neighbors, e.Node1.String())
		} else {
			nodes[e.Node2.String()].neighbors = []string{e.Node1.String()}
		}

		nodes[e.Node1.String()].capacity += e.Capacity.ToUnit(btcutil.AmountSatoshi)
		nodes[e.Node2.String()].capacity += e.Capacity.ToUnit(btcutil.AmountSatoshi)

		nodes[e.Node1.String()].channels++
		nodes[e.Node2.String()].channels++
	}

	// Add assumes to root node
	for _, c := range request.Assume {
		if _, ok := nodes[c]; !ok {
			app.Log.Printf("candidate node does not exist: %s", c)
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

	unfilteredSpan := make([]Node, 0)
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
	span := make([]Node, 0)
	for _, v := range unfilteredSpan {
		if v.capacity >= request.MinCapacity && v.channels >= request.MinChannels && v.distance >= request.MinDistance && v.updated.After(request.MinUpdated) {
			span = append(span, v)
		}
	}

	sort.Sort(sort.Reverse(sortDistance(span)))

	if len(span) < request.Limit {
		return span, nil
	}

	return span[:request.Limit], nil
}

// PrintCandidates outputs table formatted list of nodes by distance
func PrintCandidates(app App, request CandidatesRequest) error {
	nodes, err := Candidates(app, request)
	if err != nil {
		return err
	}

	tbl := table.New("Pubkey", "Alias", "Distance", "Distant Neighbors", "Capacity", "Channels", "Updated")

	for _, v := range nodes {
		tbl.AddRow(v.pubkey, v.alias, v.distance, v.distantNeigbors, v.capacity, v.channels, v.updated)
	}
	tbl.Print()

	return nil
}

// localPubkey fetches the pubkey of the local instance of lnd
func localPubkey(app App) (string, error) {
	info, err := app.Client.GetInfo(context.Background())

	if err != nil {
		return "", err
	}

	return hex.EncodeToString(info.IdentityPubkey[:]), nil
}
