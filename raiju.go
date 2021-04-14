package raiju

func Btc2sat(btc float64) int {
	return int(btc * 100000000)
}
