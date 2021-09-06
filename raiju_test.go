package raiju

import (
	"context"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/lightningnetwork/lnd/lnrpc"
	"google.golang.org/grpc"
)

const (
	rootPubkey  = "fakePubKey123"
	rootAlias   = "raiju"
	rootUpdated = "2020-01-02T15:04:05Z"
)

func TestBtcToSat(t *testing.T) {
	sats := BtcToSat(.001)
	if sats != 100000 {
		t.Fatal("btc not converted correctly to sats")
	}
}

type fakeInfoer struct {
	info *lnrpc.GetInfoResponse
}

func (f fakeInfoer) GetInfo(ctx context.Context, in *lnrpc.GetInfoRequest, opts ...grpc.CallOption) (*lnrpc.GetInfoResponse, error) {
	return f.info, nil
}

type fakeGrapher struct {
	graph *lnrpc.ChannelGraph
}

func (f fakeGrapher) DescribeGraph(ctx context.Context, in *lnrpc.ChannelGraphRequest, opts ...grpc.CallOption) (*lnrpc.ChannelGraph, error) {
	return f.graph, nil
}

func TestCandidates(t *testing.T) {
	tests := []struct {
		name    string
		graph   *lnrpc.ChannelGraph
		request CandidatesRequest
		want    []Node
	}{
		{
			name: "identity",
			graph: &lnrpc.ChannelGraph{
				Nodes: []*lnrpc.LightningNode{
					{
						PubKey:     rootPubkey,
						Alias:      rootAlias,
						LastUpdate: uint32(rootUpdatedTime(t).Unix()),
					},
				},
			},
			request: CandidatesRequest{
				MinUpdated: rootUpdatedTime(t).Add(-time.Hour * 24),
				Limit:      1,
			},
			want: []Node{
				{
					pubkey:  rootPubkey,
					alias:   rootAlias,
					updated: rootUpdatedTime(t),
				},
			},
		},
	}

	for _, tc := range tests {
		app := App{
			Infoer:  fakeInfoer{info: &lnrpc.GetInfoResponse{IdentityPubkey: rootPubkey}},
			Grapher: fakeGrapher{graph: tc.graph},
			Log:     log.New(ioutil.Discard, "", 0),
			Verbose: false,
		}

		nodes, err := Candidates(app, tc.request)

		if err != nil {
			t.Fatal("error calculating nodes by distance")
		}

		if !reflect.DeepEqual(tc.want, nodes) {
			t.Fatalf("%s candidates are incorrect\nwant: %v\ngot: %v", tc.name, tc.want, nodes)
		}
	}
}

// helper function to keep error checks out of tests
func rootUpdatedTime(t *testing.T) time.Time {
	time, err := time.Parse(time.RFC3339, rootUpdated)
	if err != nil {
		t.Fatalf("unable to parse time: %s", err)
	}
	return time
}
