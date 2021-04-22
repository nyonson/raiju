package raiju

import (
	"testing"
)

func TestBtcToSat(t *testing.T) {
	sats := BtcToSat(.001)
	if sats != 100000 {
		t.Error("btc not converted correctly to sats")
	}
}
