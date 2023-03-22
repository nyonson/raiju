package raiju

import (
	"context"
	"reflect"
	"testing"

	"github.com/nyonson/raiju/lightning"
)

const (
	rootPubkey  = "111111111112300000000000000000000000000000000000000000000000000000"
	rootAlias   = "raiju"
	rootUpdated = "2020-01-02T15:04:05Z"
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
							Pubkey: rootPubkey,
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

func TestNew(t *testing.T) {
	type args struct {
		l lightninger
	}
	tests := []struct {
		name string
		args args
		want Raiju
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := New(tt.args.l); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("New() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidityFees_High(t *testing.T) {
	type fields struct {
		standard float64
	}
	tests := []struct {
		name   string
		fields fields
		want   lightning.FeePPM
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LiquidityFees{
				standard: tt.fields.standard,
			}
			if got := l.High(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidityFees.High() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidityFees_Standard(t *testing.T) {
	type fields struct {
		standard float64
	}
	tests := []struct {
		name   string
		fields fields
		want   lightning.FeePPM
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LiquidityFees{
				standard: tt.fields.standard,
			}
			if got := l.Standard(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidityFees.Standard() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestLiquidityFees_Low(t *testing.T) {
	type fields struct {
		standard float64
	}
	tests := []struct {
		name   string
		fields fields
		want   lightning.FeePPM
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			l := LiquidityFees{
				standard: tt.fields.standard,
			}
			if got := l.Low(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("LiquidityFees.Low() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewLiquidityFees(t *testing.T) {
	type args struct {
		standard float64
	}
	tests := []struct {
		name string
		args args
		want LiquidityFees
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewLiquidityFees(tt.args.standard); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewLiquidityFees() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
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
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.s.Len(); got != tt.want {
				t.Errorf("sortDistance.Len() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestRaiju_Fees(t *testing.T) {
	type fields struct {
		l lightninger
	}
	type args struct {
		ctx  context.Context
		fees LiquidityFees
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
			r := Raiju{
				l: tt.fields.l,
			}
			if err := r.Fees(tt.args.ctx, tt.args.fees); (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Fees() error = %v, wantErr %v", err, tt.wantErr)
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
		// TODO: Add test cases.
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
		ctx           context.Context
		outChannelID  lightning.ChannelID
		lastHopPubkey string
		stepPercent   float64
		maxPercent    float64
		maxFee        lightning.FeePPM
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    float64
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := Raiju{
				l: tt.fields.l,
			}
			got, err := r.Rebalance(tt.args.ctx, tt.args.outChannelID, tt.args.lastHopPubkey, tt.args.stepPercent, tt.args.maxPercent, tt.args.maxFee)
			if (err != nil) != tt.wantErr {
				t.Errorf("Raiju.Rebalance() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("Raiju.Rebalance() = %v, want %v", got, tt.want)
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
		// TODO: Add test cases.
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
