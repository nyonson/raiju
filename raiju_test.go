package raiju

import (
	"context"
	"io/ioutil"
	"log"
	"reflect"
	"testing"
	"time"

	"github.com/lightninglabs/lndclient"
	"github.com/lightningnetwork/lnd/routing/route"
)

const (
	rootPubkey  = "111111111112300000000000000000000000000000000000000000000000000000"
	rootAlias   = "raiju"
	rootUpdated = "2020-01-02T15:04:05Z"
)

func TestBtcToSat(t *testing.T) {
	sats := BtcToSat(.001)
	if sats != 100000 {
		t.Fatal("btc not converted correctly to sats")
	}
}

type fakeClient struct {
	info  *lndclient.Info
	graph *lndclient.Graph
}

func (f fakeClient) GetInfo(ctx context.Context) (*lndclient.Info, error) {
	return f.info, nil
}

func (f fakeClient) DescribeGraph(ctx context.Context, includeUnannounced bool) (*lndclient.Graph, error) {
	return f.graph, nil
}

func TestCandidates(t *testing.T) {
	tests := []struct {
		name    string
		graph   *lndclient.Graph
		request CandidatesRequest
		want    []Node
	}{
		{
			name: "identity",
			graph: &lndclient.Graph{
				Nodes: []lndclient.Node{
					{
						PubKey:     rootVertex(t),
						Alias:      rootAlias,
						LastUpdate: rootUpdatedTime(t),
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
			Client:  fakeClient{info: &lndclient.Info{IdentityPubkey: rootVertex(t)}, graph: tc.graph},
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

// helper function to keep error checks out of tests
func rootVertex(t *testing.T) route.Vertex {
	v, err := route.NewVertexFromStr(rootPubkey)
	if err != nil {
		t.Fatalf("unable to convert vertex: %s", err)
	}

	return v
}
