package raiju

import (
	"errors"

	"github.com/nyonson/raiju/lightning"
)

// LiquidityFees for channels.
//
// Defining channel liquidity percentage based on (local capacity / total capacity).
// When liquidity is low, there is too much inbound.
// When liquidity is high, there is too much outbound.
type LiquidityFees struct {
	Thresholds []float64
	Fees       []lightning.FeePPM
	Stickiness float64
}

// Fee for channel based on its current liquidity.
func (lf LiquidityFees) Fee(channel lightning.Channel) lightning.FeePPM {
	liquidity := float64(channel.LocalBalance) / float64(channel.Capacity) * 100

	return lf.findFee(liquidity, channel.LocalFee)
}

// PotentialFee for channel based on its current liquidity.
func (lf LiquidityFees) PotentialFee(channel lightning.Channel, additionalLocal lightning.Satoshi) lightning.FeePPM {
	liquidity := float64(channel.LocalBalance+additionalLocal) / float64(channel.Capacity) * 100

	return lf.findFee(liquidity, channel.LocalFee)
}

func (lf LiquidityFees) findFee(liquidity float64, currentFee lightning.FeePPM) lightning.FeePPM {
	bucket := 0
	for bucket < len(lf.Thresholds) {
		if liquidity > lf.Thresholds[bucket] {
			break
		} else {
			bucket += 1
		}

	}

	newFee := lf.Fees[bucket]

	// apply stickiness if fee is heading in the right direction, but wanna hold on for a bit to limit gossip
	if liquidity < 50 && newFee < currentFee {
		lowBucket := 0
		for lowBucket < len(lf.Thresholds) {
			if liquidity > lf.Thresholds[lowBucket]+lf.Stickiness {
				break
			} else {
				lowBucket += 1
			}

		}

		newFee = lf.Fees[lowBucket]
	} else if liquidity >= 50 && newFee > currentFee {
		highBucket := 0
		for highBucket < len(lf.Thresholds) {
			if liquidity > lf.Thresholds[highBucket]-lf.Stickiness {
				break
			} else {
				highBucket += 1
			}

		}

		newFee = lf.Fees[highBucket]
	}

	return newFee
}

// RebalanceChannels at the far ends of the spectrum.
func (lf LiquidityFees) RebalanceChannels(channels lightning.Channels) (high lightning.Channels, low lightning.Channels) {
	for _, c := range channels {
		l := c.Liquidity()
		if l > lf.Thresholds[0] {
			high = append(high, c)
		}

		if l <= lf.Thresholds[len(lf.Thresholds)-1] {
			low = append(low, c)
		}
	}

	return high, low
}

// RebalanceFee is the max fee to use in a circular rebalance to ensure its not wasted.
func (lf LiquidityFees) RebalanceFee() lightning.FeePPM {
	return lf.Fees[len(lf.Fees)-1]
}

// NewLiquidityFees with threshold and fee validation.
func NewLiquidityFees(thresholds []float64, fees []lightning.FeePPM, stickiness float64) (LiquidityFees, error) {
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

	// ensure stickiness percent makes sense
	if stickiness > 100 {
		return LiquidityFees{}, errors.New("stickiness must be a percent")
	}

	return LiquidityFees{
		Thresholds: thresholds,
		Fees:       fees,
		Stickiness: stickiness,
	}, nil
}
