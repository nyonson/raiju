package raiju

import (
	"errors"
	"fmt"
	"os"

	"github.com/nyonson/raiju/lightning"
)

// LiquidityFees for channels.
//
// Defining channel liquidity percentage based on (local capacity / total capacity).
// When liquidity is low, there is too much inbound.
// When liquidity is high, there is too much outbound.
type LiquidityFees struct {
	thresholds []float64
	fees       []lightning.FeePPM
}

// Fee for channel based on its current liquidity.
func (lf LiquidityFees) Fee(channel lightning.Channel) lightning.FeePPM {
	liquidity := float64(channel.LocalBalance) / float64(channel.Capacity) * 100

	return lf.findFee(liquidity)
}

// PotentialFee for channel based on its current liquidity.
func (lf LiquidityFees) PotentialFee(channel lightning.Channel, additionalLocal lightning.Satoshi) lightning.FeePPM {
	liquidity := float64(channel.LocalBalance+additionalLocal) / float64(channel.Capacity) * 100

	return lf.findFee(liquidity)
}

func (lf LiquidityFees) findFee(liquidity float64) lightning.FeePPM {
	bucket := 0
	for bucket < len(lf.thresholds) {
		if liquidity > lf.thresholds[bucket] {
			break
		} else {
			bucket += 1
		}

	}

	return lf.fees[bucket]
}

// RebalanceChannels at the far ends of the spectrum.
func (lf LiquidityFees) RebalanceChannels(channels lightning.Channels) (high lightning.Channels, low lightning.Channels) {
	for _, c := range channels {
		l := c.Liquidity()
		if l > lf.thresholds[0] {
			high = append(high, c)
		}

		if l <= lf.thresholds[len(lf.thresholds)-1] {
			low = append(low, c)
		}
	}

	return high, low
}

// RebalanceFee is the max fee to use in a circular rebalance to ensure its not wasted.
func (lf LiquidityFees) RebalanceFee() lightning.FeePPM {
	return lf.fees[len(lf.fees)-1]
}

// PrintSettings to output.
func (lf LiquidityFees) PrintSettings() {
	for i := 0; i < len(lf.fees); i++ {
		if i == len(lf.fees)-1 {
			fmt.Fprintf(os.Stderr, "channels under %g%% local liquidity to %g ppm\n", lf.thresholds[i-1], lf.fees[i])
		} else if i == 0 {
			fmt.Fprintf(os.Stderr, "channels over %g%% local liquidity to %g ppm, ", lf.thresholds[i], lf.fees[i])
		} else {
			fmt.Fprintf(os.Stderr, "channels between %g%% and %g%% local liquidity to %g ppm, ", lf.thresholds[i-1], lf.thresholds[i], lf.fees[i])
		}
	}
}

// NewLiquidityFees with threshold and fee validation.
func NewLiquidityFees(thresholds []float64, fees []lightning.FeePPM) (LiquidityFees, error) {
	// ensure every bucket has a fee
	if len(thresholds)+1 != len(fees) {
		return LiquidityFees{}, errors.New("fees must have one more value than thresholds to ensure each bucket has a defined fee")

	}

	// ensure thresholds are descending
	for i := 0; i < len(thresholds)-1; i++ {
		if thresholds[i] <= thresholds[i+1] {
			return LiquidityFees{}, errors.New("thresholds must be descending")
		}
	}

	// ensure fees are ascending
	for i := 0; i < len(fees)-1; i++ {
		if fees[i] > fees[i+1] {
			return LiquidityFees{}, errors.New("fees must be ascending")
		}
	}

	return LiquidityFees{
		thresholds: thresholds,
		fees:       fees,
	}, nil
}
