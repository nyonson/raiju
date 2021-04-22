package raiju

import (
	"context"
	"fmt"
	"os"
	"sort"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/rodaine/table"
)

// BtcToSat returns the btc amount in satoshis
func BtcToSat(btc float64) int {
	return int(btc * 100000000)
}

// PrintBtcToSat prints the btc amount in satoshis to stdout
func PrintBtcToSat(btc float64) {
	fmt.Fprintln(os.Stdout, BtcToSat(btc))
}

type node struct {
	pubkey    string
	alias     string
	distance  int
	channels  int
	capacity  int64
	updated   time.Time
	neighbors []string
}

type sortSpan []node

func (s sortSpan) Less(i, j int) bool {
	if s[i].distance != s[j].distance {
		return s[i].distance < s[j].distance
	}

	if s[i].capacity != s[j].capacity {
		return s[i].capacity < s[j].capacity
	}

	return s[i].channels < s[j].channels
}

func (s sortSpan) Swap(i, j int) {
	s[i], s[j] = s[j], s[i]
}

func (s sortSpan) Len() int {
	return len(s)
}

// NodesByDistance walks the lightning network from a specific node keeping track of distance (hops)
func NodesByDistance(client lnrpc.LightningClient, pubkey string, minCapacity int64, minChannels int, minDistance int) error {
	// default root to local lnd if no key supplied
	if pubkey == "" {
		pk, err := localPubkey(client)
		if err != nil {
			return err
		}

		pubkey = pk
	}

	channelGraphRequest := lnrpc.ChannelGraphRequest{
		IncludeUnannounced: false,
	}
	channelGraph, err := client.DescribeGraph(context.Background(), &channelGraphRequest)

	if err != nil {
		return err
	}

	nodes := make(map[string]*node)

	for _, n := range channelGraph.Nodes {
		nodes[n.PubKey] = &node{pubkey: n.PubKey, alias: n.Alias, updated: time.Unix(int64(n.LastUpdate), 0)}
	}

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

	count := 1
	visited := make(map[string]bool)

	current := nodes[pubkey].neighbors
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

	unfilteredSpan := make([]node, len(nodes))
	for _, v := range nodes {
		unfilteredSpan = append(unfilteredSpan, *v)
	}

	span := make([]node, 0)
	for _, v := range unfilteredSpan {
		if v.capacity > minCapacity && v.channels > minChannels && v.distance > minDistance {
			span = append(span, v)
		}
	}

	sort.Sort(sort.Reverse(sortSpan(span)))

	tbl := table.New("Pubkey", "Alias", "Distance", "Capacity", "Channels", "Updated")

	for _, v := range span {
		tbl.AddRow(v.pubkey, v.alias, v.distance, v.capacity, v.channels, v.updated)
	}
	tbl.Print()

	return nil
}

// localPubkey fetches the pubkey of the local instance of lnd
func localPubkey(client lnrpc.LightningClient) (string, error) {
	info, err := client.GetInfo(context.Background(), &lnrpc.GetInfoRequest{})

	if err != nil {
		return "", err
	}

	return info.IdentityPubkey, nil
}
