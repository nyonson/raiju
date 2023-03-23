// Lightning Network primitives.
package lightning

import (
	"reflect"
	"testing"
	"time"
)

var (
	updated, _ = time.Parse(time.RFC3339, "2020-01-02T15:04:05Z")
)

func TestSatoshi_BTC(t *testing.T) {
	tests := []struct {
		name string
		s    Satoshi
		want float64
	}{
		{
			name: "happy conversion",
			s:    100000000,
			want: 1,
		},
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
		{
			name: "happy ppm rate",
			f:    200,
			want: 0.0002,
		},
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
	type fields struct {
		PubKey    PubKey
		Alias     string
		Updated   time.Time
		Addresses []string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name: "detect clearnet node",
			fields: fields{
				PubKey:    "A",
				Alias:     "A",
				Updated:   time.Now(),
				Addresses: []string{"123.123.123.123:12312"},
			},
			want: true,
		},
		{
			name: "detect hybrid only node",
			fields: fields{
				PubKey:    "A",
				Alias:     "A",
				Updated:   time.Now(),
				Addresses: []string{"123.123.123.123:12312", "axlvvynqvvz3f5u3dfhtsyxzeqttivnw2awas3rxniu5uvoqrlvrvgid.onion:9735"},
			},
			want: true,
		},
		{
			name: "detect tor only node",
			fields: fields{
				PubKey:    "A",
				Alias:     "A",
				Updated:   time.Now(),
				Addresses: []string{"axlvvynqvvz3f5u3dfhtsyxzeqttivnw2awas3rxniu5uvoqrlvrvgid.onion:9735"},
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			n := Node{
				PubKey:    tt.fields.PubKey,
				Alias:     tt.fields.Alias,
				Updated:   tt.fields.Updated,
				Addresses: tt.fields.Addresses,
			}
			if got := n.Clearnet(); got != tt.want {
				t.Errorf("Node.Clearnet() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_Liquidity(t *testing.T) {
	type fields struct {
		Edge          Edge
		ChannelID     ChannelID
		LocalBalance  Satoshi
		RemoteBalance Satoshi
		RemoteNode    Node
	}
	tests := []struct {
		name   string
		fields fields
		want   float64
	}{
		{
			name: "calculate liquidity",
			fields: fields{
				Edge: Edge{
					Capacity: 4,
					Node1:    "A",
					Node2:    "B",
				},
				ChannelID:     0,
				LocalBalance:  2,
				RemoteBalance: 2,
				RemoteNode: Node{
					PubKey:    "A",
					Alias:     "A",
					Updated:   time.Now(),
					Addresses: []string{"123.12.123.123:1231"},
				},
			},
			want: 50,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Channel{
				Edge:          tt.fields.Edge,
				ChannelID:     tt.fields.ChannelID,
				LocalBalance:  tt.fields.LocalBalance,
				RemoteBalance: tt.fields.RemoteBalance,
				RemoteNode:    tt.fields.RemoteNode,
			}
			if got := c.Liquidity(); got != tt.want {
				t.Errorf("Channel.Liquidity() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestChannel_LiquidityLevel(t *testing.T) {
	type fields struct {
		Edge          Edge
		ChannelID     ChannelID
		LocalBalance  Satoshi
		RemoteBalance Satoshi
		RemoteNode    Node
	}
	tests := []struct {
		name   string
		fields fields
		want   ChannelLiquidityLevel
	}{
		{
			name: "detect low liquidity",
			fields: fields{
				Edge: Edge{
					Capacity: 10,
					Node1:    "A",
					Node2:    "B",
				},
				ChannelID:     0,
				LocalBalance:  1,
				RemoteBalance: 9,
				RemoteNode: Node{
					PubKey:    "B",
					Alias:     "B",
					Updated:   time.Now(),
					Addresses: []string{"123.12.123.123:1231"},
				},
			},
			want: LowLiquidity,
		},
		{
			name: "detect standard liquidity",
			fields: fields{
				Edge: Edge{
					Capacity: 10,
					Node1:    "A",
					Node2:    "B",
				},
				ChannelID:     0,
				LocalBalance:  5,
				RemoteBalance: 5,
				RemoteNode: Node{
					PubKey:    "B",
					Alias:     "B",
					Updated:   time.Now(),
					Addresses: []string{"123.12.123.123:1231"},
				},
			},
			want: StandardLiquidity,
		},
		{
			name: "detect high liquidity",
			fields: fields{
				Edge: Edge{
					Capacity: 10,
					Node1:    "A",
					Node2:    "B",
				},
				ChannelID:     0,
				LocalBalance:  9,
				RemoteBalance: 1,
				RemoteNode: Node{
					PubKey:    "B",
					Alias:     "B",
					Updated:   time.Now(),
					Addresses: []string{"123.12.123.123:1231"},
				},
			},
			want: HighLiquidity,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := Channel{
				Edge:          tt.fields.Edge,
				ChannelID:     tt.fields.ChannelID,
				LocalBalance:  tt.fields.LocalBalance,
				RemoteBalance: tt.fields.RemoteBalance,
				RemoteNode:    tt.fields.RemoteNode,
			}
			if got := c.LiquidityLevel(); got != tt.want {
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
		{
			name: "filter low liquidity channels",
			cs: []Channel{
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "B",
					},
					ChannelID:     0,
					LocalBalance:  1,
					RemoteBalance: 9,
					RemoteNode: Node{
						PubKey:    "B",
						Alias:     "B",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "C",
					},
					ChannelID:     0,
					LocalBalance:  5,
					RemoteBalance: 5,
					RemoteNode: Node{
						PubKey:    "C",
						Alias:     "C",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "D",
					},
					ChannelID:     0,
					LocalBalance:  9,
					RemoteBalance: 1,
					RemoteNode: Node{
						PubKey:    "D",
						Alias:     "D",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
			},
			want: []Channel{
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "B",
					},
					ChannelID:     0,
					LocalBalance:  1,
					RemoteBalance: 9,
					RemoteNode: Node{
						PubKey:    "B",
						Alias:     "B",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
			},
		},
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

		{
			name: "filter high liquidity channels",
			cs: []Channel{
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "B",
					},
					ChannelID:     0,
					LocalBalance:  1,
					RemoteBalance: 9,
					RemoteNode: Node{
						PubKey:    "B",
						Alias:     "B",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "C",
					},
					ChannelID:     0,
					LocalBalance:  5,
					RemoteBalance: 5,
					RemoteNode: Node{
						PubKey:    "C",
						Alias:     "C",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "D",
					},
					ChannelID:     0,
					LocalBalance:  9,
					RemoteBalance: 1,
					RemoteNode: Node{
						PubKey:    "D",
						Alias:     "D",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
			},
			want: []Channel{
				{
					Edge: Edge{
						Capacity: 10,
						Node1:    "A",
						Node2:    "D",
					},
					ChannelID:     0,
					LocalBalance:  9,
					RemoteBalance: 1,
					RemoteNode: Node{
						PubKey:    "D",
						Alias:     "D",
						Updated:   updated,
						Addresses: []string{"123.12.123.123:1231"},
					},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tt.cs.HighLiquidity(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Channels.HighLiquidity() = %v, want %v", got, tt.want)
			}
		})
	}
}
