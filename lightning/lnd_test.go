package lightning

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/routing/route"
)

func TestLndClient_GetInfo(t *testing.T) {
	var pubKey [33]byte

	type fields struct {
		c channeler
		r router
		i invoicer
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    *Info
		wantErr bool
	}{
		{
			name: "happy get info",
			fields: fields{
				c: &channelerMock{
					GetInfoFunc: func(ctx context.Context) (*lndclient.Info, error) {
						return &lndclient.Info{
							IdentityPubkey: pubKey,
						}, nil
					},
				},
				r: &routerMock{},
				i: &invoicerMock{},
			},
			args: args{},
			want: &Info{
				PubKey: "000000000000000000000000000000000000000000000000000000000000000000",
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			got, err := l.GetInfo(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lnd.GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lnd.GetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLndClient_ListChannels(t *testing.T) {
	var localPubKey [33]byte
	var remotePubKey [33]byte = [33]byte{1}

	var remoteNode = &lndclient.Node{
		Alias:      "alias",
		Addresses:  []string{"address"},
		LastUpdate: time.Time{},
		PubKey:     remotePubKey,
	}

	var channel = lndclient.ChannelInfo{
		ChannelPoint:  "channelPoint",
		ChannelID:     1,
		Capacity:      1000,
		LocalBalance:  500,
		RemoteBalance: 500,
	}

	type fields struct {
		c       channeler
		r       router
		i       invoicer
		network string
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Channels
		wantErr bool
	}{
		{
			name: "missing node policy causes error",
			fields: fields{
				c: &channelerMock{
					GetInfoFunc: func(ctx context.Context) (*lndclient.Info, error) {
						return &lndclient.Info{
							IdentityPubkey: localPubKey,
						}, nil
					},
					GetChanInfoFunc: func(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error) {
						return &lndclient.ChannelEdge{
							ChannelPoint: "channelPoint",
							Capacity:     1000,
						}, nil
					},
					ListChannelsFunc: func(ctx context.Context, activeOnly bool, publicOnly bool) ([]lndclient.ChannelInfo, error) {
						return []lndclient.ChannelInfo{
							channel,
						}, nil
					},
				},
			},
			wantErr: true,
		},
		{
			name: "valid node policy success",
			fields: fields{
				c: &channelerMock{
					GetInfoFunc: func(ctx context.Context) (*lndclient.Info, error) {
						return &lndclient.Info{
							IdentityPubkey: localPubKey,
						}, nil
					},
					GetNodeInfoFunc: func(ctx context.Context, pubkey route.Vertex, includeChannels bool) (*lndclient.NodeInfo, error) {
						return &lndclient.NodeInfo{Node: remoteNode}, nil
					},
					GetChanInfoFunc: func(ctx context.Context, chanId uint64) (*lndclient.ChannelEdge, error) {
						return &lndclient.ChannelEdge{
							ChannelPoint: "channelPoint",
							Capacity:     1000,
							Node1:        localPubKey,
							Node2:        remotePubKey,
							Node1Policy: &lndclient.RoutingPolicy{
								TimeLockDelta:    144,
								FeeBaseMsat:      1000,
								FeeRateMilliMsat: 1,
							},
							Node2Policy: &lndclient.RoutingPolicy{
								TimeLockDelta:    144,
								FeeBaseMsat:      1000,
								FeeRateMilliMsat: 1,
							},
						}, nil
					},
					ListChannelsFunc: func(ctx context.Context, activeOnly bool, publicOnly bool) ([]lndclient.ChannelInfo, error) {
						return []lndclient.ChannelInfo{
							channel,
						}, nil
					},
				},
			},
			want: Channels{
				{
					Edge: Edge{
						Capacity: Satoshi(1000),
						Node1:    PubKey(route.Vertex(localPubKey).String()),
						Node2:    PubKey(route.Vertex(remotePubKey).String()),
					},
					ChannelID:     1,
					LocalBalance:  Satoshi(channel.LocalBalance),
					LocalFee:      1,
					RemoteBalance: Satoshi(channel.RemoteBalance),
					RemoteNode: Node{
						PubKey:    PubKey(route.Vertex(remotePubKey).String()),
						Alias:     remoteNode.Alias,
						Updated:   remoteNode.LastUpdate,
						Addresses: remoteNode.Addresses,
					},
					Private: false,
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			got, err := l.ListChannels(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LndClient.ListChannels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LndClient.ListChannels() = %v, want %v", got, tt.want)
			}
		})
	}
}
