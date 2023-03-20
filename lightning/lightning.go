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

// ChannelID for channel.
type ChannelID uint64

// Rate of fee.
func (f FeePPM) Rate() float64 {
	return float64(f) / 1000000
}

type Forward struct {
	Timestamp  time.Time
	ChannelIn  ChannelID
	ChannelOut ChannelID
}

// Node in the Lightning Network.
type Node struct {
	PubKey    string
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
	Node1    string
	Node2    string
}

// Nodes and edges of the Lightning Network.
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

// Detailed information of a payment channel between nodes.
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
	Pubkey string
}

// channeler is the minimum channel requirements from LND.
type channeler interface {
	DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error)
	ForwardingHistory(ctx context.Context,
		req lndclient.ForwardingHistoryRequest) (*lndclient.ForwardingHistoryResponse, error)
	GetChanInfo(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error)
	GetInfo(ctx context.Context) (*lndclient.Info, error)
	GetNodeInfo(ctx context.Context, pubkey route.Vertex,
		includeChannels bool) (*lndclient.NodeInfo, error)
	ListChannels(ctx context.Context, activeOnly, publicOnly bool) ([]lndclient.ChannelInfo, error)
	UpdateChanPolicy(ctx context.Context, req lndclient.PolicyUpdateRequest, chanPoint *wire.OutPoint) error
}

// router is the minimum routing requirements from LND.
type router interface {
	SendPayment(ctx context.Context, request lndclient.SendPaymentRequest) (chan lndclient.PaymentStatus, chan error, error)
}

// invoicer is the minimum routing requirements from LND.
type invoicer interface {
	AddInvoice(ctx context.Context, in *invoicesrpc.AddInvoiceData) (lntypes.Hash, string, error)
}

// New Lightning instance.
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

// GetInfo of local node.
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

// DescribeGraph of the Lightning Network.
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
			Capacity: Satoshi(e.Capacity.ToUnit(btcutil.AmountSatoshi)),
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

// GetChannel with ID.
func (l Lightning) GetChannel(ctx context.Context, channelID ChannelID) (Channel, error) {
	// returns a channel edge which doesn't have liquidity info
	ce, err := l.c.GetChanInfo(ctx, uint64(channelID))
	if err != nil {
		return Channel{}, err
	}

	local, err := l.c.GetInfo(ctx)
	if err != nil {
		return Channel{}, err
	}

	remotePubkey := ce.Node1
	if local.IdentityPubkey == remotePubkey {
		remotePubkey = ce.Node2
	}

	remote, err := l.c.GetNodeInfo(ctx, remotePubkey, false)
	if err != nil {
		return Channel{}, err
	}

	c := Channel{
		Edge: Edge{
			Capacity: Satoshi(ce.Capacity.ToUnit(btcutil.AmountSatoshi)),
			Node1:    ce.Node1.String(),
			Node2:    ce.Node2.String(),
		},
		ChannelID: ChannelID(ce.ChannelID),
		RemoteNode: Node{
			PubKey:    remote.PubKey.String(),
			Alias:     remote.Alias,
			Updated:   remote.LastUpdate,
			Addresses: remote.Addresses,
		},
	}

	// get local and remote liquidity from the list channels call
	cs, err := l.c.ListChannels(ctx, false, false)
	if err != nil {
		return Channel{}, err
	}

	for _, ci := range cs {
		if ChannelID(ci.ChannelID) == channelID {
			c.LocalBalance = Satoshi(ci.LocalBalance.ToUnit(btcutil.AmountSatoshi))
			c.RemoteBalance = Satoshi(ci.RemoteBalance.ToUnit(btcutil.AmountSatoshi))
		}
	}

	return c, nil
}

// ListChannels of local node.
func (l Lightning) ListChannels(ctx context.Context) (Channels, error) {
	channelInfos, err := l.c.ListChannels(ctx, true, true)
	if err != nil {
		return nil, err
	}

	local, err := l.c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	channels := make([]Channel, len(channelInfos))
	for i, ci := range channelInfos {
		ce, err := l.c.GetChanInfo(ctx, ci.ChannelID)
		if err != nil {
			return nil, err
		}

		remotePubkey := ce.Node1
		if local.IdentityPubkey == remotePubkey {
			remotePubkey = ce.Node2
		}

		remote, err := l.c.GetNodeInfo(ctx, remotePubkey, false)
		if err != nil {
			return nil, err
		}

		channels[i] = Channel{
			Edge: Edge{
				Capacity: Satoshi(ce.Capacity.ToUnit(btcutil.AmountSatoshi)),
				Node1:    ce.Node1.String(),
				Node2:    ce.Node2.String(),
			},
			ChannelID:     ChannelID(ci.ChannelID),
			LocalBalance:  Satoshi(ci.LocalBalance.ToUnit(btcutil.AmountSatoshi)),
			RemoteBalance: Satoshi(ci.RemoteBalance.ToUnit(btcutil.AmountSatoshi)),
			RemoteNode: Node{
				PubKey:    remote.PubKey.String(),
				Alias:     remote.Alias,
				Updated:   remote.LastUpdate,
				Addresses: remote.Addresses,
			},
		}
	}

	return channels, nil
}

// SetFees for channel with rate in ppm.
func (l Lightning) SetFees(ctx context.Context, channelID ChannelID, fee FeePPM) error {
	ce, err := l.c.GetChanInfo(ctx, uint64(channelID))
	if err != nil {
		return err
	}

	outpoint, err := decodeChannelPoint(ce.ChannelPoint)
	if err != nil {
		return err
	}

	// opinionated policy
	req := lndclient.PolicyUpdateRequest{
		BaseFeeMsat:   0,
		FeeRate:       fee.Rate(),
		TimeLockDelta: 40,
	}
	return l.c.UpdateChanPolicy(ctx, req, outpoint)
}

// AddInvoice of amount.
func (l Lightning) AddInvoice(ctx context.Context, amount Satoshi) (Invoice, error) {
	in := &invoicesrpc.AddInvoiceData{
		Value: lnwire.NewMSatFromSatoshis(btcutil.Amount(amount)),
	}
	_, invoice, err := l.i.AddInvoice(ctx, in)
	return Invoice(invoice), err
}

// SendPayment to pay for invoice.
func (l Lightning) SendPayment(ctx context.Context, invoice Invoice, outChannelID ChannelID, lastHopPubkey string, maxFee Satoshi) (Satoshi, error) {
	lhpk, err := route.NewVertexFromStr(lastHopPubkey)
	if err != nil {
		return 0, err
	}

	request := lndclient.SendPaymentRequest{
		Invoice:          string(invoice),
		MaxFee:           btcutil.Amount(maxFee),
		OutgoingChanIds:  []uint64{uint64(outChannelID)},
		LastHopPubkey:    &lhpk,
		AllowSelfPayment: true,
		Timeout:          time.Duration(60) * time.Second,
	}
	status, error, err := l.r.SendPayment(ctx, request)
	if err != nil {
		return 0, err
	}

	for {
		select {
		case s := <-status:
			if s.State == lnrpc.Payment_SUCCEEDED {
				return Satoshi(s.Fee.ToSatoshis()), nil
			}
		case e := <-error:
			return 0, fmt.Errorf("error paying invoice: %w", e)
		}
	}
}

// ForwardingHistory of node since the time give, capped at 50,000 events.
func (l Lightning) ForwardingHistory(ctx context.Context, since time.Time) ([]Forward, error) {
	maxEvents := 50000
	req := lndclient.ForwardingHistoryRequest{
		StartTime: since,
		EndTime:   time.Now(),
		MaxEvents: uint32(maxEvents),
	}
	res, err := l.c.ForwardingHistory(ctx, req)
	if err != nil {
		return nil, err
	}

	// maybe reconsider better failure method
	if len(res.Events) == maxEvents {
		return nil, errors.New("pulled too many events, lower time window")
	}

	forwards := make([]Forward, 0)
	for _, f := range res.Events {
		forward := Forward{
			Timestamp:  f.Timestamp,
			ChannelIn:  ChannelID(f.ChannelIn),
			ChannelOut: ChannelID(f.ChannelOut),
		}
		forwards = append(forwards, forward)
	}

	return forwards, nil
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
