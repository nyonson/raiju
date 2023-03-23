// Lightning Network primitives.
package lightning

import (
	"strings"
	"time"
)

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

// ChannelLiquidityLevel coarse-grained bucket based on current liquidity.
type ChannelLiquidityLevel string

// ChannelLiquidityLevels
const (
	LowLiquidity      ChannelLiquidityLevel = "low"
	StandardLiquidity ChannelLiquidityLevel = "standard"
	HighLiquidity     ChannelLiquidityLevel = "high"
)

// Channel between local and remote node.
type Channel struct {
	Edge
	ChannelID     ChannelID
	LocalBalance  Satoshi
	RemoteBalance Satoshi
	RemoteNode    Node
}

// Liquidity of the channel.
func (c Channel) Liquidity() float64 {
	return float64(c.LocalBalance) / float64(c.Capacity) * 100
}

// LiquidityLevel of the channel.
func (c Channel) LiquidityLevel() ChannelLiquidityLevel {
	// Defining channel liquidity percentage based on (local capacity / total capacity).
	// When liquidity is low, there is too much inbound.
	// When liquidity is high, there is too much outbound.
	const LOW_LIQUIDITY = 20
	const HIGH_LIQUIDITY = 80

	if c.Liquidity() < LOW_LIQUIDITY {
		return LowLiquidity
	} else if c.Liquidity() > HIGH_LIQUIDITY {
		return HighLiquidity
	}

	return StandardLiquidity
}

// Channels of node.
type Channels []Channel

// LowLiquidity channels of node.
func (cs Channels) LowLiquidity() Channels {
	ll := make(Channels, 0)

	for _, c := range cs {
		if c.LiquidityLevel() == LowLiquidity {
			ll = append(ll, c)
		}
	}

	return ll
}

// HighLiquidity channels of node.
func (cs Channels) HighLiquidity() Channels {
	hl := make(Channels, 0)

	for _, c := range cs {
		if c.LiquidityLevel() == HighLiquidity {
			hl = append(hl, c)
		}
	}

	return hl
}

// Info of a node.
type Info struct {
	PubKey PubKey
}
