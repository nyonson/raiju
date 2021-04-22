package raiju

import (
	"context"
	"fmt"
	"log"
	"os"
	"sort"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rodaine/table"
)

// App context
type App struct {
	Client  lnrpc.LightningClient
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

// node in the lightning graph network with computed properties
type node struct {
	pubkey    string
	alias     string
	distance  int
	channels  int
	capacity  int64
	updated   time.Time
	neighbors []string
}

// sortDistance sorts nodes by distance, capacity, and channels
type sortDistance []node

func (s sortDistance) Less(i, j int) bool {
	if s[i].distance != s[j].distance {
		return s[i].distance < s[j].distance
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

// NodesByDistanceRequest contains necessary info to perform sorting across the network
type NodesByDistanceRequest struct {
	// Pubkey is the key of the root node to perform crawl from
	Pubkey string
	// MinCapcity filters nodes with a minimum satoshi capacity (sum of channels)
	MinCapacity int64
	// MinChannels filters nodes with a minimum number of channels
	MinChannels int
	// MinDistance filters nodes with a minumum distance (number of hops) from the root node
	MinDistance int
	// MinUpdated filters nodes which have not been updated since time
	MinUpdated time.Time
}

// NodesByDistance walks the lightning network from a specific node keeping track of distance (hops)
func NodesByDistance(app App, request NodesByDistanceRequest) ([]node, error) {
	// default root node to local lnd if no key supplied
	if request.Pubkey == "" {
		pk, err := localPubkey(app)
		if err != nil {
			return nil, err
		}

		request.Pubkey = pk
	}

	// pull entire network graph from lnd
	channelGraphRequest := lnrpc.ChannelGraphRequest{
		IncludeUnannounced: false,
	}
	channelGraph, err := app.Client.DescribeGraph(context.Background(), &channelGraphRequest)

	if err != nil {
		return nil, err
	}

	if app.Verbose {
		app.Log.Printf("network contains %d nodes total\n", len(channelGraph.Nodes))
	}

	// initialize nodes map with static info
	nodes := make(map[string]*node, len(channelGraph.Nodes))

	for _, n := range channelGraph.Nodes {
		nodes[n.PubKey] = &node{pubkey: n.PubKey, alias: n.Alias, updated: time.Unix(int64(n.LastUpdate), 0)}
	}

	// calculate node properties based on channels: neighbors, capacity, channels
	for _, e := range channelGraph.Edges {
		if nodes[e.Node1Pub].neighbors != nil {
			nodes[e.Node1Pub].neighbors = append(nodes[e.Node1Pub].neighbors, e.Node2Pub)
		} else {
			nodes[e.Node1Pub].neighbors = []string{e.Node2Pub}
		}

		if nodes[e.Node2Pub].neighbors != nil {
			nodes[e.Node2Pub].neighbors = append(nodes[e.Node2Pub].neighbors, e.Node1Pub)
		} else {
			nodes[e.Node2Pub].neighbors = []string{e.Node1Pub}
		}

		nodes[e.Node1Pub].capacity += e.Capacity
		nodes[e.Node2Pub].capacity += e.Capacity

		nodes[e.Node1Pub].channels++
		nodes[e.Node2Pub].channels++
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

	// filter nodes by request minimums and sort them by distance
	unfilteredSpan := make([]node, len(nodes))
	for _, v := range nodes {
		unfilteredSpan = append(unfilteredSpan, *v)
	}

	span := make([]node, 0)
	for _, v := range unfilteredSpan {
		if v.capacity > request.MinCapacity && v.channels > request.MinChannels && v.distance > request.MinDistance && v.updated.After(request.MinUpdated) {
			span = append(span, v)
		}
	}

	sort.Sort(sort.Reverse(sortDistance(span)))

	return span, nil
}

// PrintNodesByDistance outputs table formatted list of nodes by distance
func PrintNodesByDistance(app App, request NodesByDistanceRequest) error {
	nodes, err := NodesByDistance(app, request)
	if err != nil {
		return err
	}

	tbl := table.New("Pubkey", "Alias", "Distance", "Capacity", "Channels", "Updated")

	for _, v := range nodes {
		tbl.AddRow(v.pubkey, v.alias, v.distance, v.capacity, v.channels, v.updated)
	}
	tbl.Print()

	return nil
}

// localPubkey fetches the pubkey of the local instance of lnd
func localPubkey(app App) (string, error) {
	info, err := app.Client.GetInfo(context.Background(), &lnrpc.GetInfoRequest{})

	if err != nil {
		return "", err
	}

	return info.IdentityPubkey, nil
}
