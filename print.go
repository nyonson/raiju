package raiju

import (
	"github.com/nyonson/raiju/lightning"
	"github.com/rodaine/table"
)

// PrintNodes in table formatted list.
func PrintNodes(nodes []RelativeNode) error {
	tbl := table.New("Pubkey", "Alias", "Distance", "Distant Neighbors", "Capacity (BTC)", "Channels", "Updated", "Addresses")

	for _, v := range nodes {
		tbl.AddRow(v.PubKey, v.Alias, v.distance, v.distantNeigbors, lightning.Satoshi(v.capacity).BTC(), v.channels, v.Updated, v.Addresses)
	}
	tbl.Print()

	return nil
}

// PrintChannels in table formatted list.
func PrintChannels(channels lightning.Channels) error {
	tbl := table.New("Channel ID", "Pubkey", "Capacity (BTC)")

	for _, c := range channels {
		tbl.AddRow(c.ChannelID, c.RemoteNode.Alias, lightning.Satoshi(c.Capacity).BTC())
	}

	tbl.Print()

	return nil
}
