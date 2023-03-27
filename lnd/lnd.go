// Hook raiju up to LND.
package lnd

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
	"github.com/lightningnetwork/lnd/lnrpc/routerrpc"
	"github.com/lightningnetwork/lnd/lntypes"
	"github.com/lightningnetwork/lnd/lnwire"
	"github.com/lightningnetwork/lnd/routing/route"
	"github.com/lightningnetwork/lnd/zpay32"
	"github.com/nyonson/raiju/lightning"
)

//go:generate gotests -w -exported .
//go:generate moq -stub -skip-ensure -out lnd_mock_test.go . channeler router invoicer

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
	SubscribeHtlcEvents(ctx context.Context) (<-chan *routerrpc.HtlcEvent, <-chan error, error)
}

// invoicer is the minimum routing requirements from LND.
type invoicer interface {
	AddInvoice(ctx context.Context, in *invoicesrpc.AddInvoiceData) (lntypes.Hash, string, error)
}

// New Lightning instance.
func New(c channeler, i invoicer, r router, network string) Lnd {
	return Lnd{
		c:       c,
		i:       i,
		r:       r,
		network: network,
	}
}

// Lnd client backed by LND node.
type Lnd struct {
	c       channeler
	r       router
	i       invoicer
	network string
}

// GetInfo of local node.
func (l Lnd) GetInfo(ctx context.Context) (*lightning.Info, error) {
	i, err := l.c.GetInfo(ctx)

	if err != nil {
		return &lightning.Info{}, err
	}

	info := lightning.Info{
		PubKey: lightning.PubKey(hex.EncodeToString(i.IdentityPubkey[:])),
	}

	return &info, nil
}

// DescribeGraph of the Lightning Network.
func (l Lnd) DescribeGraph(ctx context.Context) (*lightning.Graph, error) {
	g, err := l.c.DescribeGraph(ctx, false)

	if err != nil {
		return &lightning.Graph{}, err
	}

	// marshall nodes
	nodes := make([]lightning.Node, len(g.Nodes))
	for i, n := range g.Nodes {
		nodes[i] = lightning.Node{
			PubKey:    lightning.PubKey(n.PubKey.String()),
			Alias:     n.Alias,
			Updated:   n.LastUpdate,
			Addresses: n.Addresses,
		}
	}

	// marshall edges
	edges := make([]lightning.Edge, len(g.Edges))
	for i, e := range g.Edges {
		edges[i] = lightning.Edge{
			Capacity: lightning.Satoshi(e.Capacity.ToUnit(btcutil.AmountSatoshi)),
			Node1:    lightning.PubKey(e.Node1.String()),
			Node2:    lightning.PubKey(e.Node2.String()),
		}
	}

	graph := &lightning.Graph{
		Nodes: nodes,
		Edges: edges,
	}

	return graph, nil
}

// GetChannel with ID.
func (l Lnd) GetChannel(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error) {
	// returns a channel edge which doesn't have liquidity info
	ce, err := l.c.GetChanInfo(ctx, uint64(channelID))
	if err != nil {
		return lightning.Channel{}, err
	}

	local, err := l.c.GetInfo(ctx)
	if err != nil {
		return lightning.Channel{}, err
	}

	// figure out if which node is local and which is remote
	remotePubkey := ce.Node1
	// FeeRateMilliMsat is a weird name
	localFee := lightning.FeePPM(ce.Node2Policy.FeeRateMilliMsat)
	if local.IdentityPubkey == remotePubkey {
		remotePubkey = ce.Node2
		localFee = lightning.FeePPM(ce.Node1Policy.FeeRateMilliMsat)
	}

	remote, err := l.c.GetNodeInfo(ctx, remotePubkey, false)
	if err != nil {
		return lightning.Channel{}, err
	}

	c := lightning.Channel{
		Edge: lightning.Edge{
			Capacity: lightning.Satoshi(ce.Capacity.ToUnit(btcutil.AmountSatoshi)),
			Node1:    lightning.PubKey(ce.Node1.String()),
			Node2:    lightning.PubKey(ce.Node2.String()),
		},
		ChannelID: lightning.ChannelID(ce.ChannelID),
		LocalFee:  localFee,
		RemoteNode: lightning.Node{
			PubKey:    lightning.PubKey(remote.PubKey.String()),
			Alias:     remote.Alias,
			Updated:   remote.LastUpdate,
			Addresses: remote.Addresses,
		},
	}

	// get local and remote liquidity from the list channels call
	cs, err := l.c.ListChannels(ctx, false, false)
	if err != nil {
		return lightning.Channel{}, err
	}

	for _, ci := range cs {
		if lightning.ChannelID(ci.ChannelID) == channelID {
			c.LocalBalance = lightning.Satoshi(ci.LocalBalance.ToUnit(btcutil.AmountSatoshi))
			c.RemoteBalance = lightning.Satoshi(ci.RemoteBalance.ToUnit(btcutil.AmountSatoshi))
		}
	}

	return c, nil
}

// ListChannels of local node.
func (l Lnd) ListChannels(ctx context.Context) (lightning.Channels, error) {
	channelInfos, err := l.c.ListChannels(ctx, true, true)
	if err != nil {
		return nil, err
	}

	local, err := l.c.GetInfo(ctx)
	if err != nil {
		return nil, err
	}

	channels := make([]lightning.Channel, len(channelInfos))
	for i, ci := range channelInfos {
		ce, err := l.c.GetChanInfo(ctx, ci.ChannelID)
		if err != nil {
			return nil, err
		}

		// figure out if which node is local and which is remote
		remotePubkey := ce.Node1
		localFee := lightning.FeePPM(ce.Node2Policy.FeeRateMilliMsat)
		if local.IdentityPubkey == remotePubkey {
			remotePubkey = ce.Node2
			localFee = lightning.FeePPM(ce.Node1Policy.FeeRateMilliMsat)
		}

		remote, err := l.c.GetNodeInfo(ctx, remotePubkey, false)
		if err != nil {
			return nil, err
		}

		channels[i] = lightning.Channel{
			Edge: lightning.Edge{
				Capacity: lightning.Satoshi(ce.Capacity.ToUnit(btcutil.AmountSatoshi)),
				Node1:    lightning.PubKey(ce.Node1.String()),
				Node2:    lightning.PubKey(ce.Node2.String()),
			},
			ChannelID:     lightning.ChannelID(ci.ChannelID),
			LocalBalance:  lightning.Satoshi(ci.LocalBalance.ToUnit(btcutil.AmountSatoshi)),
			LocalFee:      localFee,
			RemoteBalance: lightning.Satoshi(ci.RemoteBalance.ToUnit(btcutil.AmountSatoshi)),
			RemoteNode: lightning.Node{
				PubKey:    lightning.PubKey(remote.PubKey.String()),
				Alias:     remote.Alias,
				Updated:   remote.LastUpdate,
				Addresses: remote.Addresses,
			},
		}
	}

	return channels, nil
}

// SetFees for channel with rate in ppm.
func (l Lnd) SetFees(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error {
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
func (l Lnd) AddInvoice(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error) {
	in := &invoicesrpc.AddInvoiceData{
		Value: lnwire.NewMSatFromSatoshis(btcutil.Amount(amount)),
	}
	_, invoice, err := l.i.AddInvoice(ctx, in)
	return lightning.Invoice(invoice), err
}

// SendPayment to pay for invoice.
func (l Lnd) SendPayment(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.FeePPM) (lightning.Satoshi, error) {
	lhpk, err := route.NewVertexFromStr(string(lastHopPubKey))
	if err != nil {
		return 0, err
	}

	// decode invoice to get amount in millisats and calculate max fee from ppm
	params, err := lndclient.Network(l.network).ChainParams()
	if err != nil {
		return 0, err
	}
	i, err := zpay32.Decode(string(invoice), params)
	if err != nil {
		return 0, err
	}
	maxFeeMsat := uint64(float64(*i.MilliSat) * maxFee.Rate())

	request := lndclient.SendPaymentRequest{
		Invoice:          string(invoice),
		MaxFeeMsat:       lnwire.MilliSatoshi(maxFeeMsat),
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
				return lightning.Satoshi(s.Fee.ToSatoshis()), nil
			}
		case e := <-error:
			return 0, fmt.Errorf("error paying invoice: %w", e)
		}
	}
}

// SubscribeChannelUpdates signals when a channel's liquidity changes.
func (l Lnd) SubscribeChannelUpdates(ctx context.Context) (<-chan lightning.Channels, <-chan error, error) {
	cc := make(chan lightning.Channels)
	ec := make(chan error)

	htlcs, errors, err := l.r.SubscribeHtlcEvents(ctx)
	if err != nil {
		return nil, nil, err
	}

	// translate htlc events into channels
	go func() {
		for {
			select {
			case h := <-htlcs:
				channels := make(lightning.Channels, 0)

				if h.GetIncomingChannelId() != 0 {
					c, err := l.GetChannel(ctx, lightning.ChannelID(h.GetIncomingChannelId()))
					if err != nil {
						ec <- err
						break
					}
					channels = append(channels, c)
				}

				if h.GetOutgoingChannelId() != 0 {
					c, err := l.GetChannel(ctx, lightning.ChannelID(h.GetOutgoingChannelId()))
					if err != nil {
						ec <- err
						break
					}
					channels = append(channels, c)
				}

				cc <- channels
			case err = <-errors:
				ec <- err
			}
		}
	}()

	return cc, ec, nil
}

// ForwardingHistory of node since the time give, capped at 50,000 events.
func (l Lnd) ForwardingHistory(ctx context.Context, since time.Time) ([]lightning.Forward, error) {
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

	forwards := make([]lightning.Forward, 0)
	for _, f := range res.Events {
		forward := lightning.Forward{
			Timestamp:  f.Timestamp,
			ChannelIn:  lightning.ChannelID(f.ChannelIn),
			ChannelOut: lightning.ChannelID(f.ChannelOut),
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
