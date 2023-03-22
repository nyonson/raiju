package raiju

import (
	"github.com/nyonson/raiju/lightning"
	"github.com/rodaine/table"
)

// printNodes in table formatted list.
func printNodes(nodes []RelativeNode) error {
	tbl := table.New("Pubkey", "Alias", "Distance", "Distant Neighbors", "Capacity (BTC)", "Channels", "Updated", "Addresses")

	for _, v := range nodes {
		tbl.AddRow(v.PubKey, v.Alias, v.distance, v.distantNeigbors, lightning.Satoshi(v.capacity).BTC(), v.channels, v.Updated, v.Addresses)
	}
	tbl.Print()

	return nil
}

// printChannels in table formatted list.
func printChannels(channels lightning.Channels) error {
	tbl := table.New("Channel ID", "Alias", "Capacity (BTC)")

	for _, c := range channels {
		tbl.AddRow(c.ChannelID, c.RemoteNode.Alias, lightning.Satoshi(c.Capacity).BTC())
	}

	tbl.Print()

	return nil
}
