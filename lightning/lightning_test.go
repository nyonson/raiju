// Wrap lightning network node implementations.
package lightning

import (
	"context"
	"reflect"
	"testing"
	"time"
)

func TestSatoshi_BTC(t *testing.T) {
	tests := []struct {
		name string
		s    Satoshi
		want float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.BTC(); got != tt.want {
				t.Errorf("Satoshi.BTC() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFeePPM_Rate(t *testing.T) {
	tests := []struct {
		name string
		f    FeePPM
		want float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.f.Rate(); got != tt.want {
				t.Errorf("FeePPM.Rate() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNode_Clearnet(t *testing.T) {
	tests := []struct {
		name string
		n    Node
		want bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.n.Clearnet(); got != tt.want {
				t.Errorf("Node.Clearnet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_Liquidity(t *testing.T) {
	tests := []struct {
		name string
		c    Channel
		want float64
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.Liquidity(); got != tt.want {
				t.Errorf("Channel.Liquidity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_LiquidityLevel(t *testing.T) {
	tests := []struct {
		name string
		c    Channel
		want ChannelLiquidityLevel
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.c.LiquidityLevel(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Channel.LiquidityLevel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannels_LowLiquidity(t *testing.T) {
	tests := []struct {
		name string
		cs   Channels
		want Channels
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.LowLiquidity(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Channels.LowLiquidity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannels_HighLiquidity(t *testing.T) {
	tests := []struct {
		name string
		cs   Channels
		want Channels
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.HighLiquidity(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Channels.HighLiquidity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNew(t *testing.T) {
	type args struct {
		c channeler
		i invoicer
		r router
	}
	tests := []struct {
		name string
		args args
		want Lightning
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.c, tt.args.i, tt.args.r); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLightning_GetInfo(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		want    *Info
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.GetInfo(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lightning.GetInfo() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lightning.GetInfo() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLightning_DescribeGraph(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		want    *Graph
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.DescribeGraph(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lightning.DescribeGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lightning.DescribeGraph() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLightning_GetChannel(t *testing.T) {
	type args struct {
		ctx       context.Context
		channelID ChannelID
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		want    Channel
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.GetChannel(tt.args.ctx, tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lightning.GetChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lightning.GetChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLightning_ListChannels(t *testing.T) {
	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		want    Channels
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.ListChannels(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lightning.ListChannels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lightning.ListChannels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLightning_SetFees(t *testing.T) {
	type args struct {
		ctx       context.Context
		channelID ChannelID
		fee       FeePPM
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.l.SetFees(tt.args.ctx, tt.args.channelID, tt.args.fee); (err != nil) != tt.wantErr {
				t.Errorf("Lightning.SetFees() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLightning_AddInvoice(t *testing.T) {
	type args struct {
		ctx    context.Context
		amount Satoshi
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		want    Invoice
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.AddInvoice(tt.args.ctx, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lightning.AddInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lightning.AddInvoice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLightning_SendPayment(t *testing.T) {
	type args struct {
		ctx           context.Context
		invoice       Invoice
		outChannelID  ChannelID
		lastHopPubkey string
		maxFee        Satoshi
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		want    Satoshi
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.SendPayment(tt.args.ctx, tt.args.invoice, tt.args.outChannelID, tt.args.lastHopPubkey, tt.args.maxFee)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lightning.SendPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lightning.SendPayment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLightning_ForwardingHistory(t *testing.T) {
	type args struct {
		ctx   context.Context
		since time.Time
	}
	tests := []struct {
		name    string
		l       Lightning
		args    args
		want    []Forward
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := tt.l.ForwardingHistory(tt.args.ctx, tt.args.since)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lightning.ForwardingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lightning.ForwardingHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}
