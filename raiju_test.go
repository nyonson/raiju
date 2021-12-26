package raiju

import (
	"context"
	"reflect"
	"testing"
	"time"

	"git.sr.ht/~yonson/raiju/lightning"
)

const (
	rootPubkey  = "111111111112300000000000000000000000000000000000000000000000000000"
	rootAlias   = "raiju"
	rootUpdated = "2020-01-02T15:04:05Z"
)

type fakeClient struct {
	getInfo       func(ctx context.Context) (*lightning.Info, error)
	describeGraph func(ctx context.Context) (*lightning.Graph, error)
}

func (f fakeClient) GetInfo(ctx context.Context) (*lightning.Info, error) {
	if f.getInfo != nil {
		return f.getInfo(ctx)
	}

	return &lightning.Info{
		Pubkey: rootPubkey,
	}, nil
}

func (f fakeClient) DescribeGraph(ctx context.Context) (*lightning.Graph, error) {
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

func TestBtcToSat(t *testing.T) {
	sats := BtcToSat(.001)
	if sats != 100000 {
		t.Fatal("btc not converted correctly to sats")
	}
}

func TestRaiju_Candidates(t *testing.T) {
	type fields struct {
		client client
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
				client: fakeClient{},
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
				client: tt.fields.client,
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
