// Wrap lightning network node implementations.
package lightning

import (
	"context"
	"encoding/hex"
	"time"

	"github.com/btcsuite/btcutil"
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

// Info of a node.
type Info struct {
	Pubkey string
}

// lnder is the minimal requirements from LND.
type lnder interface {
	GetInfo(ctx context.Context) (*lndclient.Info, error)
	DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error)
}

// New client.
func New(lnd lnder) LndClient {
	return LndClient{
		lnd: lnd,
	}
}

// Lightning client backed by LND node.
type LndClient struct {
	lnd lnder
}

func (lc LndClient) GetInfo(ctx context.Context) (*Info, error) {
	i, err := lc.lnd.GetInfo(ctx)

	if err != nil {
		return &Info{}, err
	}

	info := Info{
		Pubkey: hex.EncodeToString(i.IdentityPubkey[:]),
	}

	return &info, nil
}

func (lc LndClient) DescribeGraph(ctx context.Context) (*Graph, error) {
	g, err := lc.lnd.DescribeGraph(ctx, false)

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
