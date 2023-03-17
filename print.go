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
