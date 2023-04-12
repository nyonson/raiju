package view

import (
	"context"
	"strconv"
	"time"

	"github.com/nyonson/raiju"
	"github.com/nyonson/raiju/lightning"
	"github.com/rivo/tview"
)

func ViewCandidates(ctx context.Context, r raiju.Raiju) (*tview.Flex, error) {
	container := tview.NewFlex()
	container.SetBorder(true).SetTitle("Candidates")

	table := tview.NewTable().SetSelectable(true, false)

	form := tview.NewForm().
		AddInputField("Capacity", "10000000", 12, nil, nil).
		AddInputField("Distance", "2", 4, nil, nil).
		AddInputField("Distant Neighbors", "0", 4, nil, nil)

	form.SetBorder(true).SetTitle("Min Filters")

	form.AddButton("Refresh", func() {
		table.Clear()

		minCapacity, _ := strconv.Atoi(form.GetFormItem(0).(*tview.InputField).GetText())
		minDistance, _ := strconv.Atoi(form.GetFormItem(1).(*tview.InputField).GetText())
		minDistantNeighbors, _ := strconv.Atoi(form.GetFormItem(2).(*tview.InputField).GetText())

		request := raiju.CandidatesRequest{
			MinCapacity:         lightning.Satoshi(minCapacity),
			MinChannels:         1,
			MinDistance:         int64(minDistance),
			MinDistantNeighbors: int64(minDistantNeighbors),
			MinUpdated:          time.Now().Add(-2 * 24 * time.Hour),
			Limit:               200,
			Clearnet:            true,
		}
		nodes, err := r.Candidates(ctx, request)
		if err != nil {
			return
		}

		// headers
		// would like to show aliases, but double width emoji are the worst
		// tview might handle it better if this gets in: https://github.com/mattn/go-runewidth/pull/63
		table.SetCellSimple(0, 0, "PubKey")
		table.SetCellSimple(0, 1, "Distance")
		table.SetCellSimple(0, 2, "Distant Neighbors")
		// always show header row
		table.SetFixed(1, 0)

		// content
		row := 1
		for _, n := range nodes {
			table.SetCellSimple(row, 0, string(n.PubKey))
			table.SetCellSimple(row, 1, strconv.FormatInt(n.Distance, 10))
			table.SetCellSimple(row, 2, strconv.FormatInt(n.DistantNeigbors, 10))
			row++
		}

		table.ScrollToBeginning()
	})

	container.AddItem(form, 36, 1, true)
	container.AddItem(table, 0, 4, false)

	return container, nil
}

func ViewChannels(ctx context.Context, r raiju.Raiju) (*tview.Flex, error) {
	container := tview.NewFlex()
	container.SetBorder(true).SetTitle("Channels")

	return container, nil
}
