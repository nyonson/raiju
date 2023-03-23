// Hook raiju up to LND.
package lnd

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/nyonson/raiju/lightning"
)

func TestNew(t *testing.T) {
	type args struct {
		c channeler
		i invoicer
		r router
	}
	tests := []struct {
		name string
		args args
		want Lnd
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

func TestLnd_GetInfo(t *testing.T) {
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
		want    *lightning.Info
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
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

func TestLnd_DescribeGraph(t *testing.T) {
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
		want    *lightning.Graph
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			got, err := l.DescribeGraph(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lnd.DescribeGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lnd.DescribeGraph() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLnd_GetChannel(t *testing.T) {
	type fields struct {
		c channeler
		r router
		i invoicer
	}
	type args struct {
		ctx       context.Context
		channelID lightning.ChannelID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    lightning.Channel
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			got, err := l.GetChannel(tt.args.ctx, tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lnd.GetChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lnd.GetChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLnd_ListChannels(t *testing.T) {
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
		want    lightning.Channels
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			got, err := l.ListChannels(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lnd.ListChannels() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lnd.ListChannels() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLnd_SetFees(t *testing.T) {
	type fields struct {
		c channeler
		r router
		i invoicer
	}
	type args struct {
		ctx       context.Context
		channelID lightning.ChannelID
		fee       lightning.FeePPM
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			if err := l.SetFees(tt.args.ctx, tt.args.channelID, tt.args.fee); (err != nil) != tt.wantErr {
				t.Errorf("Lnd.SetFees() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLnd_AddInvoice(t *testing.T) {
	type fields struct {
		c channeler
		r router
		i invoicer
	}
	type args struct {
		ctx    context.Context
		amount lightning.Satoshi
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    lightning.Invoice
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			got, err := l.AddInvoice(tt.args.ctx, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lnd.AddInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lnd.AddInvoice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLnd_SendPayment(t *testing.T) {
	type fields struct {
		c channeler
		r router
		i invoicer
	}
	type args struct {
		ctx           context.Context
		invoice       lightning.Invoice
		outChannelID  lightning.ChannelID
		lastHopPubKey lightning.PubKey
		maxFee        lightning.Satoshi
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    lightning.Satoshi
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			got, err := l.SendPayment(tt.args.ctx, tt.args.invoice, tt.args.outChannelID, tt.args.lastHopPubKey, tt.args.maxFee)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lnd.SendPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lnd.SendPayment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLnd_ForwardingHistory(t *testing.T) {
	type fields struct {
		c channeler
		r router
		i invoicer
	}
	type args struct {
		ctx   context.Context
		since time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []lightning.Forward
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := Lnd{
				c: tt.fields.c,
				r: tt.fields.r,
				i: tt.fields.i,
			}
			got, err := l.ForwardingHistory(tt.args.ctx, tt.args.since)
			if (err != nil) != tt.wantErr {
				t.Errorf("Lnd.ForwardingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Lnd.ForwardingHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}
