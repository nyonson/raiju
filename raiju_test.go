package raiju

import (
	"context"
	"reflect"
	"testing"
	"time"

	"github.com/nyonson/raiju/lightning"
)

const (
	rootPubkey  = "111111111112300000000000000000000000000000000000000000000000000000"
	rootAlias   = "raiju"
	rootUpdated = "2020-01-02T15:04:05Z"
)

type fakeLightninger struct {
	getInfo       func(ctx context.Context) (*lightning.Info, error)
	describeGraph func(ctx context.Context) (*lightning.Graph, error)
	listChannels  func(ctx context.Context) (lightning.Channels, error)
	setFees       func(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error
}

func (f fakeLightninger) AddInvoice(ctx context.Context, amount lightning.Satoshi) (lightning.Invoice, error) {
	return "", nil
}

func (f fakeLightninger) ForwardingHistory(ctx context.Context, since time.Time) ([]lightning.Forward, error) {
	return []lightning.Forward{}, nil
}

func (f fakeLightninger) SendPayment(ctx context.Context, invoice lightning.Invoice, outChannelID lightning.ChannelID, lastHopPubkey string, maxFee lightning.Satoshi) (lightning.Satoshi, error) {
	return 0, nil
}

func (f fakeLightninger) GetChannel(ctx context.Context, channelID lightning.ChannelID) (lightning.Channel, error) {
	return lightning.Channel{}, nil
}

func (f fakeLightninger) GetInfo(ctx context.Context) (*lightning.Info, error) {
	if f.getInfo != nil {
		return f.getInfo(ctx)
	}

	return &lightning.Info{
		Pubkey: rootPubkey,
	}, nil
}

func (f fakeLightninger) DescribeGraph(ctx context.Context) (*lightning.Graph, error) {
	if f.describeGraph != nil {
		return f.describeGraph(ctx)
	}

	n := lightning.Node{
		PubKey:  rootPubkey,
		Alias:   rootAlias,
		Updated: time.Time{},
	}

	return &lightning.Graph{
		Nodes: []lightning.Node{n},
	}, nil
}

func (f fakeLightninger) ListChannels(ctx context.Context) (lightning.Channels, error) {
	if f.listChannels != nil {
		return f.listChannels(ctx)
	}

	return nil, nil
}

func (f fakeLightninger) SetFees(ctx context.Context, channelID lightning.ChannelID, fee lightning.FeePPM) error {
	if f.setFees != nil {
		return f.setFees(ctx, channelID, fee)
	}

	return nil
}

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
				l: fakeLightninger{},
			},
			args: args{
				request: CandidatesRequest{},
			},
			want:    []RelativeNode{},
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
