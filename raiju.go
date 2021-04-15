package raiju

import (
	"fmt"
	"os"
)

// Btc2sat returns the btc amount in satoshis
func Btc2sat(btc float64) int {
	return int(btc * 100000000)
}

// PrintBtc2sat prints the btc amount in satoshis to stdout
func PrintBtc2sat(btc float64) {
	fmt.Fprintln(os.Stdout, Btc2sat(btc))
}
