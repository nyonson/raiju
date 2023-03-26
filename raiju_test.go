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
					distance:        3,
					distantNeigbors: 0,
					channels:        1,
					capacity:        1,
					neighbors:       []lightning.PubKey{pubKeyC},
				},
				{
					Node: lightning.Node{
						PubKey:    "C",
						Alias:     "C",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
					distance:        2,
					distantNeigbors: 1,
					channels:        2,
					capacity:        2,
					neighbors:       []lightning.PubKey{pubKeyB, pubKeyD},
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
					distance:        3,
					distantNeigbors: 0,
					channels:        2,
					capacity:        2,
					neighbors:       []lightning.PubKey{pubKeyC, pubKeyE},
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
					distance:        0,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					distance:        1,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
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
					distance:        1,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					distance:        1,
					distantNeigbors: 1,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
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
					distance:        1,
					distantNeigbors: 1,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					distance:        1,
					distantNeigbors: 1,
					channels:        0,
					capacity:        1,
					neighbors:       []lightning.PubKey{},
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
					distance:        1,
					distantNeigbors: 1,
					channels:        0,
					capacity:        1,
					neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					distance:        1,
					distantNeigbors: 1,
					channels:        1,
					capacity:        1,
					neighbors:       []lightning.PubKey{},
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
					distance:        0,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
				},
				{
					Node: lightning.Node{
						PubKey:    pubKey,
						Alias:     alias,
						Updated:   updated,
						Addresses: []string{},
					},
					distance:        0,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
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
					distance:        0,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
				},
				{
					Node:            lightning.Node{},
					distance:        0,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
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
					distance:        0,
					distantNeigbors: 0,
					channels:        0,
					capacity:        0,
					neighbors:       []lightning.PubKey{},
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

func TestRaiju_RebalanceAll(t *testing.T) {
	type fields struct {
		l lightninger
	}
	type args struct {
		ctx         context.Context
		stepPercent float64
		maxPercent  float64
		maxFee      lightning.FeePPM
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		{
			name: "happy rebalance all",
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
				stepPercent: 1,
				maxPercent:  5,
				maxFee:      10,
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Raiju{
				l: tt.fields.l,
			}
			if err := r.RebalanceAll(tt.args.ctx, tt.args.stepPercent, tt.args.maxPercent, tt.args.maxFee); (err != nil) != tt.wantErr {
				t.Errorf("Raiju.RebalanceAll() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestRaiju_Rebalance(t *testing.T) {
	type fields struct {
		l lightninger
	}
	type args struct {
		ctx           context.Context
		outChannelID  lightning.ChannelID
		lastHopPubKey lightning.PubKey
		stepPercent   float64
		maxPercent    float64
		maxFee        lightning.FeePPM
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		want1   lightning.Satoshi
		wantErr bool
	}{
		{
			name: "rebalance to max percent",
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
					SendPaymentFunc: func(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubKey lightning.PubKey, maxFee lightning.FeePPM) (lightning.Satoshi, error) {
						return 1, nil
					},
				},
			},
			args: args{
				outChannelID:  0,
				lastHopPubKey: pubKey,
				stepPercent:   1,
				maxPercent:    5,
				maxFee:        10,
			},
			want:    5,
			want1:   5,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Raiju{
				l: tt.fields.l,
			}
			got, got1, err := r.Rebalance(tt.args.ctx, tt.args.outChannelID, tt.args.lastHopPubKey, tt.args.stepPercent, tt.args.maxPercent, tt.args.maxFee)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Rebalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Raiju.Rebalance() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("Raiju.Rebalance() got1 = %v, want %v", got1, tt.want1)
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
		ctx    context.Context
		daemon bool
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
					thresholds: []float64{80, 20},
					fees:       []lightning.FeePPM{5, 10, 100},
				},
			},
			args: args{
				daemon: false,
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
			got, err := r.Fees(tt.args.ctx, tt.args.daemon)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Fees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Raiju.Fees() = %v, want %v", got, tt.want)
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
					thresholds: []float64{80, 20},
					fees:       []lightning.FeePPM{},
				},
			},
			want: Raiju{
				l: nil,
				f: LiquidityFees{
					thresholds: []float64{80, 20},
					fees:       []lightning.FeePPM{},
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
