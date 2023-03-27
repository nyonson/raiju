package raiju

import (
	"reflect"
	"testing"

	"github.com/nyonson/raiju/lightning"
)

func TestLiquidityFees_Fee(t *testing.T) {
	type fields struct {
		thresholds []float64
		fees       []lightning.FeePPM
	}
	type args struct {
		channel lightning.Channel
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   lightning.FeePPM
	}{
		{
			name: "grab fee based on liquidity",
			fields: fields{
				thresholds: []float64{80, 20},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			args: args{
				channel: lightning.Channel{
					Edge: lightning.Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "B",
					},
					ChannelID:     1,
					LocalBalance:  1,
					LocalFee:      50,
					RemoteBalance: 9,
					RemoteNode: lightning.Node{
						PubKey:    pubKeyB,
						Alias:     "B",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
				},
			},
			want: lightning.FeePPM(500),
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf := LiquidityFees{
				thresholds: tt.fields.thresholds,
				fees:       tt.fields.fees,
			}
			if got := lf.Fee(tt.args.channel); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidityFees.Fee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidityFees_PotentialFee(t *testing.T) {
	type fields struct {
		thresholds []float64
		fees       []lightning.FeePPM
	}
	type args struct {
		channel         lightning.Channel
		additionalLocal lightning.Satoshi
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   lightning.FeePPM
	}{
		{
			name: "grab fee based on potential liquidity",
			fields: fields{
				thresholds: []float64{80, 20},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			args: args{
				additionalLocal: lightning.Satoshi(3),
				channel: lightning.Channel{
					Edge: lightning.Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "B",
					},
					ChannelID:     1,
					LocalBalance:  1,
					LocalFee:      50,
					RemoteBalance: 9,
					RemoteNode: lightning.Node{
						PubKey:    pubKeyB,
						Alias:     "B",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
				},
			},
			want: 50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf := LiquidityFees{
				thresholds: tt.fields.thresholds,
				fees:       tt.fields.fees,
			}
			if got := lf.PotentialFee(tt.args.channel, tt.args.additionalLocal); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidityFees.PotentialFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidityFees_RebalanceChannels(t *testing.T) {
	type fields struct {
		thresholds []float64
		fees       []lightning.FeePPM
	}
	type args struct {
		channels lightning.Channels
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantHigh lightning.Channels
		wantLow  lightning.Channels
	}{
		{
			name: "get the highest and lowest liquidity channels",
			fields: fields{
				thresholds: []float64{80, 20},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			args: args{
				channels: lightning.Channels{
					{
						Edge: lightning.Edge{
							Capacity: 10,
							Node1:    "A",
							Node2:    "B",
						},
						ChannelID:     1,
						LocalBalance:  9,
						LocalFee:      50,
						RemoteBalance: 1,
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
							Node1:    "A",
							Node2:    "C",
						},
						ChannelID:     2,
						LocalBalance:  5,
						LocalFee:      50,
						RemoteBalance: 5,
						RemoteNode: lightning.Node{
							PubKey:    pubKeyC,
							Alias:     "C",
							Updated:   updated,
							Addresses: []string{clearnetAddress},
						},
					},
					{
						Edge: lightning.Edge{
							Capacity: 10,
							Node1:    "A",
							Node2:    "D",
						},
						ChannelID:     3,
						LocalBalance:  1,
						LocalFee:      50,
						RemoteBalance: 9,
						RemoteNode: lightning.Node{
							PubKey:    pubKeyD,
							Alias:     "D",
							Updated:   updated,
							Addresses: []string{clearnetAddress},
						},
					},
				},
			},
			wantHigh: []lightning.Channel{
				{
					Edge: lightning.Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "B",
					},
					ChannelID:     1,
					LocalBalance:  9,
					LocalFee:      50,
					RemoteBalance: 1,
					RemoteNode: lightning.Node{
						PubKey:    pubKeyB,
						Alias:     "B",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
				},
			},
			wantLow: []lightning.Channel{
				{
					Edge: lightning.Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "D",
					},
					ChannelID:     3,
					LocalBalance:  1,
					LocalFee:      50,
					RemoteBalance: 9,
					RemoteNode: lightning.Node{
						PubKey:    pubKeyD,
						Alias:     "D",
						Updated:   updated,
						Addresses: []string{clearnetAddress},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf := LiquidityFees{
				thresholds: tt.fields.thresholds,
				fees:       tt.fields.fees,
			}
			gotHigh, gotLow := lf.RebalanceChannels(tt.args.channels)
			if !reflect.DeepEqual(gotHigh, tt.wantHigh) {
				t.Errorf("LiquidityFees.RebalanceChannels() gotHigh = %v, want %v", gotHigh, tt.wantHigh)
			}
			if !reflect.DeepEqual(gotLow, tt.wantLow) {
				t.Errorf("LiquidityFees.RebalanceChannels() gotLow = %v, want %v", gotLow, tt.wantLow)
			}
		})
	}
}

func TestLiquidityFees_RebalanceFee(t *testing.T) {
	type fields struct {
		thresholds []float64
		fees       []lightning.FeePPM
	}
	tests := []struct {
		name   string
		fields fields
		want   lightning.FeePPM
	}{
		{
			name: "get lowest liquidity fee",
			fields: fields{
				thresholds: []float64{80, 20},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			want: 500,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lf := LiquidityFees{
				thresholds: tt.fields.thresholds,
				fees:       tt.fields.fees,
			}
			if got := lf.RebalanceFee(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidityFees.RebalanceFee() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLiquidityFees(t *testing.T) {
	type args struct {
		thresholds []float64
		fees       []lightning.FeePPM
	}
	tests := []struct {
		name    string
		args    args
		want    LiquidityFees
		wantErr bool
	}{
		{
			name: "happy case",
			args: args{
				thresholds: []float64{80, 20},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			want: LiquidityFees{
				thresholds: []float64{80, 20},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			wantErr: false,
		},
		{
			name: "missing fees or thresholds",
			args: args{
				thresholds: []float64{80, 60, 20},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			wantErr: true,
		},
		{
			name: "thresholds must descend",
			args: args{
				thresholds: []float64{80, 85},
				fees:       []lightning.FeePPM{5, 50, 500},
			},
			wantErr: true,
		},
		{
			name: "fees must ascend",
			args: args{
				thresholds: []float64{80, 20},
				fees:       []lightning.FeePPM{5, 2, 500},
			},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := NewLiquidityFees(tt.args.thresholds, tt.args.fees)
			if (err != nil) != tt.wantErr {
				t.Errorf("NewLiquidityFees() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLiquidityFees() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidityFees_PrintSettings(t *testing.T) {
	type fields struct {
		thresholds []float64
		fees       []lightning.FeePPM
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// No tests
	}
	for _, tt := range tests {
		t.Run(tt.name, func(_ *testing.T) {
			lf := LiquidityFees{
				thresholds: tt.fields.thresholds,
				fees:       tt.fields.fees,
			}
			lf.PrintSettings()
		})
	}
}
