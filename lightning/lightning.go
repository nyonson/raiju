// Wrap lightning network node implementations.
package lightning

import (
	"context"
	"encoding/hex"
	"errors"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/btcsuite/btcd/btcutil"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/lnrpc"
	"github.com/lightningnetwork/lnd/lnrpc/invoicesrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/routing/route"
)

// Node in the Lightning Network.
type Node struct {
	PubKey    string
	Alias     string
	Updated   time.Time
	Addresses []string
}

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
	Capacity btcutil.Amount
	Node1    string
	Node2    string
}

// Nodes and edges of the Lightning Network.
type Graph struct {
	Nodes []Node
	Edges []Edge
}

// Detailed information of a payment channel between nodes.
type Channel struct {
	Edge
	ChannelID uint64
	Local     btcutil.Amount
	Remote    btcutil.Amount
}

// Info of a node.
type Info struct {
	Pubkey string
}

// channeler is the minimum channel requirements from LND.
type channeler interface {
	DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error)
	GetChanInfo(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error)
	GetInfo(ctx context.Context) (*lndclient.Info, error)
	ListChannels(ctx context.Context, activeOnly, publicOnly bool) ([]lndclient.ChannelInfo, error)
	UpdateChanPolicy(ctx context.Context, req lndclient.PolicyUpdateRequest, chanPoint *wire.OutPoint) error
}

// router is the minimum routing requirements from LND.
type router interface {
	SendPayment(ctx context.Context, request lndclient.SendPaymentRequest) (chan lndclient.PaymentStatus, chan error, error)
}

type invoicer interface {
	AddInvoice(ctx context.Context, in *invoicesrpc.AddInvoiceData) (lntypes.Hash, string, error)
}

// New client.
func New(c channeler, i invoicer, r router) Lightning {
	return Lightning{
		c: c,
		i: i,
		r: r,
	}
}

// Lightning client backed by LND node.
type Lightning struct {
	c channeler
	r router
	i invoicer
}

func (l Lightning) GetInfo(ctx context.Context) (*Info, error) {
	i, err := l.c.GetInfo(ctx)

	if err != nil {
		return &Info{}, err
	}

	info := Info{
		Pubkey: hex.EncodeToString(i.IdentityPubkey[:]),
	}

	return &info, nil
}

func (l Lightning) DescribeGraph(ctx context.Context) (*Graph, error) {
	g, err := l.c.DescribeGraph(ctx, false)

	if err != nil {
		return &Graph{}, err
	}

	// marshall nodes
	nodes := make([]Node, len(g.Nodes))
	for i, n := range g.Nodes {
		nodes[i] = Node{
			PubKey:    n.PubKey.String(),
			Alias:     n.Alias,
			Updated:   n.LastUpdate,
			Addresses: n.Addresses,
		}
	}

	// marshall edges
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

func (l Lightning) GetChannel(ctx context.Context, channelID uint64) (Channel, error) {
	// lazy, but letting list channels handle the data joining and marshaling
	channels, err := l.ListChannels(ctx)
	if err != nil {
		return Channel{}, err
	}

	for _, c := range channels {
		if c.ChannelID == channelID {
			return c, nil
		}
	}

	return Channel{}, errors.New("no channel with that ID")
}

func (l Lightning) ListChannels(ctx context.Context) ([]Channel, error) {
	channelInfos, err := l.c.ListChannels(ctx, true, true)

	if err != nil {
		return nil, err
	}

	channels := make([]Channel, len(channelInfos))
	for i, ci := range channelInfos {
		ce, err := l.c.GetChanInfo(ctx, ci.ChannelID)
		if err != nil {
			return nil, err
		}

		channels[i] = Channel{
			Edge:      Edge{Capacity: ci.Capacity, Node1: ce.Node1.String(), Node2: ce.Node2.String()},
			ChannelID: ci.ChannelID,
			Local:     ci.LocalBalance,
			Remote:    ci.RemoteBalance,
		}
	}

	return channels, nil
}

func (l Lightning) SetFees(ctx context.Context, channelID uint64, fee int64) error {
	ce, err := l.c.GetChanInfo(ctx, channelID)
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
	return l.c.UpdateChanPolicy(ctx, req, outpoint)
}

func (l Lightning) AddInvoice(ctx context.Context, amount int64) (string, error) {
	in := &invoicesrpc.AddInvoiceData{
		Value: lnwire.NewMSatFromSatoshis(btcutil.Amount(amount)),
	}
	_, invoice, err := l.i.AddInvoice(ctx, in)
	return invoice, err
}

func (l Lightning) SendPayment(ctx context.Context, invoice string, outChannelID uint64, lastHopPubkey string, maxFee int64) error {
	lhpk, err := route.NewVertexFromStr(lastHopPubkey)
	if err != nil {
		return err
	}

	request := lndclient.SendPaymentRequest{
		Invoice:          invoice,
		MaxFee:           btcutil.Amount(maxFee),
		OutgoingChanIds:  []uint64{outChannelID},
		LastHopPubkey:    &lhpk,
		AllowSelfPayment: true,
	}
	status, error, err := l.r.SendPayment(ctx, request)
	if err != nil {
		return err
	}

	for {
		select {
		case s := <-status:
			if s.State == lnrpc.Payment_SUCCEEDED {
				return nil
			}
		case e := <-error:
			return fmt.Errorf("error paying invoice: %w", e)
		}
	}
}

// ChannelLiquidities in coarse buckets based on current state.
func ChannelLiquidities(channels []Channel) (lowLiquidityChannels []Channel, standardLiquidityChannels []Channel, highLiquidityChannels []Channel) {
	// Defining channel liquidity percentage based on (local capacity / total capacity).
	// When liquidity is low, there is too much inbound.
	// When liquidity is high, there is too much outbound.
	const LOW_LIQUIDITY = 20
	const HIGH_LIQUIDITY = 80

	for _, c := range channels {
		liquidity := c.Local.ToUnit(btcutil.AmountSatoshi) / (c.Local.ToUnit(btcutil.AmountSatoshi) + c.Remote.ToUnit(btcutil.AmountSatoshi)) * 100

		if liquidity < LOW_LIQUIDITY {
			lowLiquidityChannels = append(lowLiquidityChannels, c)
		} else if liquidity > HIGH_LIQUIDITY {
			highLiquidityChannels = append(highLiquidityChannels, c)
		} else {
			standardLiquidityChannels = append(standardLiquidityChannels, c)
		}
	}

	return lowLiquidityChannels, standardLiquidityChannels, highLiquidityChannels
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
