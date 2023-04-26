// Lightning Network primitives.
package lightning

import (
	"strings"
	"time"
)

//go:generate gotests -w -exported .

// Satoshi unit of bitcoin.
type Satoshi int64

// BTC value of Satoshi.
func (s Satoshi) BTC() float64 {
	return float64(s) / 100000000
}

// FeePPM is the channel fee in part per million.
type FeePPM float64

// Invoice for lightning payment.
type Invoice string

// PubKey of node.
type PubKey string

// ChannelID for channel.
type ChannelID uint64

// Rate of fee.
func (f FeePPM) Rate() float64 {
	return float64(f) / 1000000
}

// Forward routing event.
type Forward struct {
	Timestamp  time.Time
	ChannelIn  ChannelID
	ChannelOut ChannelID
}

// Node in the Lightning Network.
type Node struct {
	PubKey    PubKey
	Alias     string
	Updated   time.Time
	Addresses []string
}

// Clearnet is true if node has a clearnet address.
func (n Node) Clearnet() bool {
	clearnet := false

	for _, a := range n.Addresses {
		// simple check filtering tor addresses
		if !strings.Contains(a, "onion") {
			clearnet = true
		}

	}

	return clearnet
}

// Edge between nodes in the Lightning Network.
type Edge struct {
	Capacity Satoshi
	Node1    PubKey
	Node2    PubKey
}

// Graph of nodes and edges of the Lightning Network.
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Channel between local and remote node.
type Channel struct {
	Edge
	ChannelID     ChannelID
	LocalBalance  Satoshi
	LocalFee      FeePPM
	RemoteBalance Satoshi
	RemoteNode    Node
	Private       bool
}

// Liquidity percent of the channel that is local.
func (c Channel) Liquidity() float64 {
	return float64(c.LocalBalance) / float64(c.Capacity) * 100
}

// Channels of node.
type Channels []Channel

// Info of a node.
type Info struct {
	PubKey PubKey
}
