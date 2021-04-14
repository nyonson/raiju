package raiju

import (
	"testing"
)

func TestBtc2Sat(t *testing.T) {
	sats := Btc2sat(.001)
	if sats != 100000 {
		t.Error("btc not converted correctly to sats")
	}
}
