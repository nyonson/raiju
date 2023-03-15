package raiju

import (
	"github.com/rodaine/table"
)

// PrintNodes in table formatted list.
func PrintNodes(nodes []RelativeNode) error {
	tbl := table.New("Pubkey", "Alias", "Distance", "Distant Neighbors", "Capacity (BTC)", "Channels", "Updated", "Addresses")

	for _, v := range nodes {
		tbl.AddRow(v.PubKey, v.Alias, v.distance, v.distantNeigbors, SatsToBtc(v.capacity), v.channels, v.Updated, v.Addresses)
	}
	tbl.Print()

	return nil
}
