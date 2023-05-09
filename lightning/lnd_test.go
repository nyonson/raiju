package lightning

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/lightninglabs/lndclient"
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

func TestNewLndClient(t *testing.T) {
	type args struct {
		c       channeler
		i       invoicer
		r       router
		network string
	}
	tests := []struct {
		name string
		args args
		want LndClient
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLndClient(tt.args.c, tt.args.i, tt.args.r, tt.args.network); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLndClient() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLndClient_DescribeGraph(t *testing.T) {
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
		want    *Graph
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			got, err := l.DescribeGraph(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LndClient.DescribeGraph() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LndClient.DescribeGraph() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLndClient_ListChannels(t *testing.T) {
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
		// TODO: Add test cases.
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

func TestLndClient_SetFees(t *testing.T) {
	type fields struct {
		c       channeler
		r       router
		i       invoicer
		network string
	}
	type args struct {
		ctx       context.Context
		channelID ChannelID
		fee       FeePPM
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
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			if err := l.SetFees(tt.args.ctx, tt.args.channelID, tt.args.fee); (err != nil) != tt.wantErr {
				t.Errorf("LndClient.SetFees() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestLndClient_AddInvoice(t *testing.T) {
	type fields struct {
		c       channeler
		r       router
		i       invoicer
		network string
	}
	type args struct {
		ctx    context.Context
		amount Satoshi
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Invoice
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			got, err := l.AddInvoice(tt.args.ctx, tt.args.amount)
			if (err != nil) != tt.wantErr {
				t.Errorf("LndClient.AddInvoice() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LndClient.AddInvoice() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLndClient_SendPayment(t *testing.T) {
	type fields struct {
		c       channeler
		r       router
		i       invoicer
		network string
	}
	type args struct {
		ctx           context.Context
		invoice       Invoice
		outChannelID  ChannelID
		lastHopPubKey PubKey
		maxFee        FeePPM
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Satoshi
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			got, err := l.SendPayment(tt.args.ctx, tt.args.invoice, tt.args.outChannelID, tt.args.lastHopPubKey, tt.args.maxFee)
			if (err != nil) != tt.wantErr {
				t.Errorf("LndClient.SendPayment() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("LndClient.SendPayment() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLndClient_SubscribeChannelUpdates(t *testing.T) {
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
		want    <-chan Channels
		want1   <-chan error
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			got, got1, err := l.SubscribeChannelUpdates(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("LndClient.SubscribeChannelUpdates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LndClient.SubscribeChannelUpdates() got = %v, want %v", got, tt.want)
			}
			if !reflect.DeepEqual(got1, tt.want1) {
				t.Errorf("LndClient.SubscribeChannelUpdates() got1 = %v, want %v", got1, tt.want1)
			}
		})
	}
}

func TestLndClient_ForwardingHistory(t *testing.T) {
	type fields struct {
		c       channeler
		r       router
		i       invoicer
		network string
	}
	type args struct {
		ctx   context.Context
		since time.Time
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []Forward
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			got, err := l.ForwardingHistory(tt.args.ctx, tt.args.since)
			if (err != nil) != tt.wantErr {
				t.Errorf("LndClient.ForwardingHistory() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LndClient.ForwardingHistory() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLndClient_GetChannel(t *testing.T) {
	type fields struct {
		c       channeler
		r       router
		i       invoicer
		network string
	}
	type args struct {
		ctx       context.Context
		channelID ChannelID
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    Channel
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LndClient{
				c:       tt.fields.c,
				r:       tt.fields.r,
				i:       tt.fields.i,
				network: tt.fields.network,
			}
			got, err := l.GetChannel(tt.args.ctx, tt.args.channelID)
			if (err != nil) != tt.wantErr {
				t.Errorf("LndClient.GetChannel() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LndClient.GetChannel() = %v, want %v", got, tt.want)
			}
		})
	}
}
