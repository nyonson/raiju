// Wrap lightning network node implementations.
package lightning

import (
	"context"
	"encoding/hex"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightninglabs/lndclient"
)

// Node in the lightning network.
type Node struct {
	PubKey  string
	Alias   string
	Updated time.Time
}

// Edge between nodes in the lightning network.
type Edge struct {
	Capacity btcutil.Amount
	Node1    string
	Node2    string
}

// Nodes and edges of the lightning network.
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Detailed information of a channel between nodes.
type Channel struct {
	Edge
	ChannelID uint64
	Local     btcutil.Amount
	Remote    btcutil.Amount
	Fee       int64
}

// Info of a node.
type Info struct {
	Pubkey string
}

// lnder is the minimal requirements from LND.
type lnder interface {
	DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error)
	GetChanInfo(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error)
	GetInfo(ctx context.Context) (*lndclient.Info, error)
	ListChannels(ctx context.Context, activeOnly, publicOnly bool) ([]lndclient.ChannelInfo, error)
	UpdateChanPolicy(ctx context.Context, req lndclient.PolicyUpdateRequest, chanPoint *wire.OutPoint) error
}

// New client.
func New(lnd lnder) Lightning {
	return Lightning{
		lnd: lnd,
	}
}

// Lightning client backed by LND node.
type Lightning struct {
	lnd lnder
}

func (l Lightning) GetInfo(ctx context.Context) (*Info, error) {
	i, err := l.lnd.GetInfo(ctx)

	if err != nil {
		return &Info{}, err
	}

	info := Info{
		Pubkey: hex.EncodeToString(i.IdentityPubkey[:]),
	}

	return &info, nil
}

func (l Lightning) DescribeGraph(ctx context.Context) (*Graph, error) {
	g, err := l.lnd.DescribeGraph(ctx, false)

	if err != nil {
		return &Graph{}, err
	}

	// marshall nodes
	nodes := make([]Node, len(g.Nodes))
	for i, n := range g.Nodes {
		nodes[i] = Node{
			PubKey:  n.PubKey.String(),
			Alias:   n.Alias,
			Updated: n.LastUpdate,
		}
	}

	// marsholl edges
	edges := make([]Edge, len(g.Edges))
	for i, e := range g.Edges {
		edges[i] = Edge{
			Capacity: e.Capacity,
			Node1:    e.Node1.String(),
			Node2:    e.Node2.String(),
		}
	}

	graph := &Graph{
		Nodes: nodes,
		Edges: edges,
	}

	return graph, nil
}

func (l Lightning) ListChannels(ctx context.Context) ([]Channel, error) {
	info, err := l.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	channelInfos, err := l.lnd.ListChannels(ctx, true, true)

	if err != nil {
		return nil, err
	}

	channels := make([]Channel, len(channelInfos))
	for i, ci := range channelInfos {
		ce, err := l.lnd.GetChanInfo(ctx, ci.ChannelID)
		if err != nil {
			return nil, err
		}

		// get fee of local node
		fee := ce.Node1Policy.FeeRateMilliMsat
		if ce.Node2.String() == info.Pubkey {
			fee = ce.Node2Policy.FeeRateMilliMsat
		}

		channels[i] = Channel{
			Edge:      Edge{Capacity: ci.Capacity, Node1: ce.Node1.String(), Node2: ce.Node2.String()},
			ChannelID: ci.ChannelID,
			Local:     ci.LocalBalance,
			Remote:    ci.RemoteBalance,
			Fee:       fee,
		}
	}

	return channels, nil
}

func (l Lightning) SetFees(ctx context.Context, channelID uint64, fee int) error {
	ce, err := l.lnd.GetChanInfo(ctx, channelID)
	if err != nil {
		return err
	}

	outpoint, err := decodeChannelPoint(ce.ChannelPoint)
	if err != nil {
		return err
	}

	feerate := float64(fee) / 1000000

	// opinionated policy
	req := lndclient.PolicyUpdateRequest{
		BaseFeeMsat:   0,
		FeeRate:       feerate,
		TimeLockDelta: 40,
	}
	return l.lnd.UpdateChanPolicy(ctx, req, outpoint)
}

func decodeChannelPoint(cp string) (*wire.OutPoint, error) {
	split := strings.SplitN(cp, ":", 2)

	hash, err := chainhash.NewHashFromStr(split[0])
	if err != nil {
		return nil, err
	}

	index, err := strconv.ParseUint(split[1], 10, 32)
	if err != nil {
		return nil, err
	}

	return &wire.OutPoint{
		Hash:  *hash,
		Index: uint32(index),
	}, nil
}
