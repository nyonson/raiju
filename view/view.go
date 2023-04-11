package view

import (
	"context"
	"strconv"
	"time"

	"github.com/nyonson/raiju"
	"github.com/rivo/tview"
)

func ViewCandidates(ctx context.Context, r raiju.Raiju) (*tview.Flex, error) {
	flex := tview.NewFlex()

	table := tview.NewTable().SetSelectable(true, false)
	table.SetBorder(true).SetTitle("Candidates")

	button := tview.NewButton("Refresh")

	flex.AddItem(table, 0, 4, false)
	flex.AddItem(button, 0, 1, true)

	request := raiju.CandidatesRequest{
		MinCapacity:         1000000,
		MinChannels:         1,
		MinDistance:         2,
		MinDistantNeighbors: 0,
		MinUpdated:          time.Now().Add(-2 * 24 * time.Hour),
		Limit:               200,
		Clearnet:            true,
	}
	nodes, err := r.Candidates(ctx, request)
	if err != nil {
		return nil, err
	}

	// headers
	table.SetCellSimple(0, 0, "PubKey")
	table.SetCellSimple(0, 2, "Distance")
	table.SetCellSimple(0, 3, "Distant Neighbors")

	// content
	row := 1
	for _, n := range nodes {
		table.SetCellSimple(row, 0, string(n.PubKey))
		table.SetCellSimple(row, 2, strconv.FormatInt(n.Distance, 10))
		table.SetCellSimple(row, 3, strconv.FormatInt(n.DistantNeigbors, 10))
		row++
	}

	return flex, nil
}
