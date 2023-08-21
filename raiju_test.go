package raiju

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/nyonson/raiju/lightning"
)

const (
	pubKey          = lightning.PubKey("111111111112300000000000000000000000000000000000000000000000000000")
	pubKeyA         = lightning.PubKey("A")
	pubKeyB         = lightning.PubKey("B")
	pubKeyC         = lightning.PubKey("C")
	pubKeyD         = lightning.PubKey("D")
	pubKeyE         = lightning.PubKey("E")
	pubKeyF         = lightning.PubKey("F")
	pubKeyG         = lightning.PubKey("G")
	alias           = "raiju"
	clearnetAddress = "44.127.188.136:9735"
)

var (
	updated, _ = time.Parse(time.RFC3339, "2020-01-02T15:04:05Z")
)

func TestRaiju_Candidates(t *testing.T) {
	type fields struct {
		l lightninger
	}
	type args struct {
		ctx     context.Context
		request CandidatesRequest
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []RelativeNode
		wantErr bool
	}{
		{
			name: "zero graph",
			fields: fields{
				l: &lightningerMock{
					DescribeGraphFunc: func(ctx context.Context) (*lightning.Graph, error) {
						return &lightning.Graph{
							Nodes: []lightning.Node{},
							Edges: []lightning.Edge{},
						}, nil
					},
					GetInfoFunc: func(ctx context.Context) (*lightning.Info, error) {
						return &lightning.Info{
							PubKey: pubKey,
						}, nil
					},
				},
			},
			args: args{
				request: CandidatesRequest{},
			},
			want:    []RelativeNode{},
			wantErr: false,
		},
		{
			name: "min distance request should filter out close nodes",
			fields: fields{
				l: &lightningerMock{
					DescribeGraphFunc: func(ctx context.Context) (*lightning.Graph, error) {
						// a linear network (A) <=> (B) <=> (C) <=> (D)
						return &lightning.Graph{
							Nodes: []lightning.Node{
								{
									PubKey:    pubKeyA,
									Alias:     "A",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyB,
									Alias:     "B",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyC,
									Alias:     "C",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyD,
									Alias:     "D",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
							},
							Edges: []lightning.Edge{
								{
									Capacity: 1,
									Node1:    pubKeyA,
									Node2:    pubKeyB,
								},
								{
									Capacity: 1,
									Node1:    pubKeyB,
									Node2:    pubKeyC,
								},
								{
									Capacity: 1,
									Node1:    pubKeyC,
									Node2:    pubKeyD,
								},
							},
						}, nil
					},
					GetInfoFunc: func(ctx context.Context) (*lightning.Info, error) {
						return &lightning.Info{
							PubKey: pubKey,
						}, nil
					},
				},
			},
			args: args{
				request: CandidatesRequest{
					PubKey:              lightning.PubKey("A"),
					MinCapacity:         1,
					MinChannels:         1,
					MinDistance:         2,
					MinDistantNeighbors: 0,
					MinUpdated:          updated.Add(time.Hour * -3),
					Assume:              []lightning.PubKey{},
					Limit:               10,
					Clearnet:            true,
				},
			},
			want: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    pubKeyD,
						Alias:     "D",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
					Distance:        3,
					DistantNeigbors: 0,
					Channels:        1,
					Capacity:        1,
					Neighbors:       []lightning.PubKey{pubKeyC},
				},
				{
					Node: lightning.Node{
						PubKey:    "C",
						Alias:     "C",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
					Distance:        2,
					DistantNeigbors: 1,
					Channels:        2,
					Capacity:        2,
					Neighbors:       []lightning.PubKey{pubKeyB, pubKeyD},
				},
			},
			wantErr: false,
		},
		{
			name: "assume should look like channel and change candidates",
			fields: fields{
				l: &lightningerMock{
					DescribeGraphFunc: func(ctx context.Context) (*lightning.Graph, error) {
						// a linear network (A) <=> (B) <=> (C) <=> (D) <=> (E) <=> (F) <=> (G)
						return &lightning.Graph{
							Nodes: []lightning.Node{
								{
									PubKey:    pubKeyA,
									Alias:     "A",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyB,
									Alias:     "B",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyC,
									Alias:     "C",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyD,
									Alias:     "D",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyE,
									Alias:     "E",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyF,
									Alias:     "F",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
								{
									PubKey:    pubKeyG,
									Alias:     "G",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
							},
							Edges: []lightning.Edge{
								{
									Capacity: 1,
									Node1:    pubKeyA,
									Node2:    pubKeyB,
								},
								{
									Capacity: 1,
									Node1:    pubKeyB,
									Node2:    pubKeyC,
								},
								{
									Capacity: 1,
									Node1:    pubKeyC,
									Node2:    pubKeyD,
								},
								{
									Capacity: 1,
									Node1:    pubKeyD,
									Node2:    pubKeyE,
								},
								{
									Capacity: 1,
									Node1:    pubKeyE,
									Node2:    pubKeyF,
								},
								{
									Capacity: 1,
									Node1:    pubKeyF,
									Node2:    pubKeyG,
								},
							},
						}, nil
					},
					GetInfoFunc: func(ctx context.Context) (*lightning.Info, error) {
						return &lightning.Info{
							PubKey: pubKey,
						}, nil
					},
				},
			},
			args: args{
				request: CandidatesRequest{
					PubKey:              lightning.PubKey("A"),
					MinCapacity:         1,
					MinChannels:         1,
					MinDistance:         3,
					MinDistantNeighbors: 0,
					MinUpdated:          updated.Add(time.Hour * -3),
					Assume:              []lightning.PubKey{pubKeyF},
					Limit:               10,
					Clearnet:            true,
				},
			},
			want: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    "D",
						Alias:     "D",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
					Distance:        3,
					DistantNeigbors: 0,
					Channels:        2,
					Capacity:        2,
					Neighbors:       []lightning.PubKey{pubKeyC, pubKeyE},
				},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Raiju{
				l: tt.fields.l,
			}
			got, err := r.Candidates(tt.args.ctx, tt.args.request)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Candidates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Raiju.Candidates() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sortDistance_Less(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		s    sortDistance
		args args
		want bool
	}{
		{
			name: "sort by distance first",
			s: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        0,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        1,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
			},

			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "sort by neighbors second",
			s: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        1,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        1,
					DistantNeigbors: 1,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
			},

			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "sort by capacity third",
			s: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        1,
					DistantNeigbors: 1,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        1,
					DistantNeigbors: 1,
					Channels:        0,
					Capacity:        1,
					Neighbors:       []lightning.PubKey{},
				},
			},

			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "sort by channels forth",
			s: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        1,
					DistantNeigbors: 1,
					Channels:        0,
					Capacity:        1,
					Neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        1,
					DistantNeigbors: 1,
					Channels:        1,
					Capacity:        1,
					Neighbors:       []lightning.PubKey{},
				},
			},

			args: args{
				i: 0,
				j: 1,
			},
			want: true,
		},
		{
			name: "all things same is a false",
			s: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        0,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        0,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
			},

			args: args{
				i: 0,
				j: 1,
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Less(tt.args.i, tt.args.j); got != tt.want {
				t.Errorf("sortDistance.Less() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_sortDistance_Swap(t *testing.T) {
	type args struct {
		i int
		j int
	}
	tests := []struct {
		name string
		s    sortDistance
		args args
	}{
		{
			name: "happy swap",
			s: []RelativeNode{
				{
					Node:            lightning.Node{},
					Distance:        0,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
				{
					Node:            lightning.Node{},
					Distance:        0,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
			},
			args: args{
				i: 0,
				j: 1,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			tt.s.Swap(tt.args.i, tt.args.j)
		})
	}
}

func Test_sortDistance_Len(t *testing.T) {
	tests := []struct {
		name string
		s    sortDistance
		want int
	}{
		{
			name: "happy length",
			s: []RelativeNode{
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					Distance:        0,
					DistantNeigbors: 0,
					Channels:        0,
					Capacity:        0,
					Neighbors:       []lightning.PubKey{},
				},
			},
			want: 1,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Len(); got != tt.want {
				t.Errorf("sortDistance.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaiju_Reaper(t *testing.T) {
	type fields struct {
		l lightninger
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    lightning.Channels
		wantErr bool
	}{
		{
			name: "detect no recent forwards",
			fields: fields{
				l: &lightningerMock{
					ForwardingHistoryFunc: func(ctx context.Context, since time.Time) ([]lightning.Forward, error) {
						return []lightning.Forward{}, nil
					},
					ListChannelsFunc: func(ctx context.Context) (lightning.Channels, error) {
						return lightning.Channels{{
							Edge: lightning.Edge{
								Capacity: 0,
								Node1:    "",
								Node2:    "",
							},
							ChannelID:     0,
							LocalBalance:  0,
							RemoteBalance: 0,
							RemoteNode: lightning.Node{
								PubKey:    pubKey,
								Alias:     alias,
								Updated:   updated,
								Addresses: []string{},
							},
						}}, nil
					},
				},
			},
			args: args{},
			want: []lightning.Channel{{
				Edge: lightning.Edge{
					Capacity: 0,
					Node1:    "",
					Node2:    "",
				},
				ChannelID:     0,
				LocalBalance:  0,
				RemoteBalance: 0,
				RemoteNode: lightning.Node{
					PubKey:    pubKey,
					Alias:     alias,
					Updated:   updated,
					Addresses: []string{},
				},
			},
			},
			wantErr: false,
		},
		{
			name: "detect recent forwards",
			fields: fields{
				l: &lightningerMock{
					ForwardingHistoryFunc: func(ctx context.Context, since time.Time) ([]lightning.Forward, error) {
						return []lightning.Forward{{
							Timestamp:  time.Time{},
							ChannelIn:  0,
							ChannelOut: 1,
						}}, nil
					},
					ListChannelsFunc: func(ctx context.Context) (lightning.Channels, error) {
						return lightning.Channels{{
							Edge: lightning.Edge{
								Capacity: 0,
								Node1:    "",
								Node2:    "",
							},
							ChannelID:     0,
							LocalBalance:  0,
							RemoteBalance: 0,
							RemoteNode: lightning.Node{
								PubKey:    pubKey,
								Alias:     alias,
								Updated:   updated,
								Addresses: []string{},
							},
						}}, nil
					},
				},
			},
			args:    args{},
			want:    []lightning.Channel{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Raiju{
				l: tt.fields.l,
			}
			got, err := r.Reaper(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Reaper() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Raiju.Reaper() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaiju_Rebalance(t *testing.T) {
	type fields struct {
		l lightninger
	}
	type args struct {
		ctx        context.Context
		maxPercent float64
		maxFee     lightning.FeePPM
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[lightning.ChannelID]float64
		wantErr bool
	}{
		{
			name: "rebalance all only one channel",
			fields: fields{
				l: &lightningerMock{
					AddInvoiceFunc: func(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error) {
						return lightning.Invoice(""), nil
					},
					GetChannelFunc: func(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error) {
						return lightning.Channel{
							Edge:          lightning.Edge{},
							ChannelID:     0,
							LocalBalance:  0,
							RemoteBalance: 0,
							RemoteNode:    lightning.Node{},
						}, nil
					},
					GetInfoFunc: func(ctx context.Context) (*lightning.Info, error) {
						return &lightning.Info{
							PubKey: pubKey,
						}, nil
					},
					ListChannelsFunc: func(ctx context.Context) (lightning.Channels, error) {
						return lightning.Channels{}, nil
					},
					SendPaymentFunc: func(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.FeePPM) (lightning.Satoshi, error) {
						return 0, nil
					},
				},
			},
			args: args{
				maxPercent: 5,
				maxFee:     lightning.FeePPM(1024),
			},
			want:    map[lightning.ChannelID]float64{},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Raiju{
				l: tt.fields.l,
			}
			got, err := r.Rebalance(tt.args.ctx, tt.args.maxPercent, tt.args.maxFee)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Rebalance() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Raiju.Rebalance() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaiju_Fees(t *testing.T) {
	type fields struct {
		l lightninger
		f LiquidityFees
	}
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    map[lightning.ChannelID]lightning.FeePPM
		wantErr bool
	}{
		{
			name: "update channel fees when necessary",
			fields: fields{
				l: &lightningerMock{
					ListChannelsFunc: func(ctx context.Context) (lightning.Channels, error) {
						return lightning.Channels{
							{
								Edge: lightning.Edge{
									Capacity: 10,
									Node1:    pubKeyA,
									Node2:    pubKeyB,
								},
								ChannelID:     1,
								LocalBalance:  1,
								LocalFee:      10,
								RemoteBalance: 9,
								RemoteNode: lightning.Node{
									PubKey:    pubKeyB,
									Alias:     "B",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
							},
							{
								Edge: lightning.Edge{
									Capacity: 10,
									Node1:    pubKeyA,
									Node2:    pubKeyC,
								},
								ChannelID:     2,
								LocalBalance:  5,
								LocalFee:      10,
								RemoteBalance: 5,
								RemoteNode: lightning.Node{
									PubKey:    pubKeyC,
									Alias:     "C",
									Updated:   updated,
									Addresses: []string{clearnetAddress},
								},
							},
						}, nil
					},
					SetFeesFunc: func(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error {
						return nil
					},
				},
				f: LiquidityFees{
					Thresholds: []float64{80, 20},
					Fees:       []lightning.FeePPM{5, 10, 100},
				},
			},
			args: args{
				ctx: context.Background(),
			},
			want: map[lightning.ChannelID]lightning.FeePPM{
				lightning.ChannelID(1): lightning.FeePPM(100),
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Raiju{
				l: tt.fields.l,
				f: tt.fields.f,
			}

			uc, ec, err := r.Fees(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Fees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			select {
			case got := <-uc:
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("Raiju.Fees() = %v, want %v", got, tt.want)
				}
			case err := <-ec:
				t.Errorf("Raiju.Fees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		l lightninger
		r LiquidityFees
	}
	tests := []struct {
		name string
		args args
		want Raiju
	}{
		{
			name: "happy init",
			args: args{
				l: nil,
				r: LiquidityFees{
					Thresholds: []float64{80, 20},
					Fees:       []lightning.FeePPM{},
				},
			},
			want: Raiju{
				l: nil,
				f: LiquidityFees{
					Thresholds: []float64{80, 20},
					Fees:       []lightning.FeePPM{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.l, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}
