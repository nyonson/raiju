package raiju

import (
	"github.com/rodaine/table"
)

// PrintNodes outputs table formatted list of nodes.
func PrintNodes(nodes []RelativeNode) error {
	tbl := table.New("Pubkey", "Alias", "Distance", "Distant Neighbors", "Capacity", "Channels", "Updated")

	for _, v := range nodes {
		tbl.AddRow(v.PubKey, v.Alias, v.distance, v.distantNeigbors, v.capacity, v.channels, v.Updated)
	}
	tbl.Print()

	return nil
}
