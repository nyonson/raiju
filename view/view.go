package view

import (
	"github.com/nyonson/raiju"
	"github.com/nyonson/raiju/lightning"
	"github.com/rodaine/table"
)

// PrintNodes in table formatted list.
func PrintNodes(nodes []raiju.RelativeNode) error {
	tbl := table.New("Pubkey", "Alias", "Distance", "Distant Neighbors", "Capacity (BTC)", "Channels", "Updated", "Addresses")

	for _, v := range nodes {
		tbl.AddRow(v.PubKey, v.Alias, v.Distance, v.DistantNeigbors, lightning.Satoshi(v.Capacity).BTC(), v.Channels, v.Updated, v.Addresses)
	}
	tbl.Print()

	return nil
}

// PrintChannels in table formatted list.
func PrintChannels(channels lightning.Channels) error {
	tbl := table.New("Channel ID", "Alias", "Capacity (BTC)")

	for _, c := range channels {
		tbl.AddRow(c.ChannelID, c.RemoteNode.Alias, lightning.Satoshi(c.Capacity).BTC())
	}

	tbl.Print()

	return nil
}

// PrintSettings to output.
func PrintFees(lf raiju.LiquidityFees) error {
	tbl := table.New("Local Liquidity Threshold Percent", "Fee PPM")

	for i := 0; i < len(lf.Thresholds); i++ {
		tbl.AddRow(lf.Thresholds[i], lf.Fees[i])
	}

	tbl.AddRow(0, lf.Fees[len(lf.Fees)-1])

	tbl.Print()

	return nil
}
