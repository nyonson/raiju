package lightning

import (
	"context"
	"reflect"
	"testing"

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
