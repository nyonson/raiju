package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/nyonson/raiju"
	"github.com/peterbourgon/ff/v3/ffcli"
)

func main() {
	cmdLog := log.New(os.Stderr, "raiju: ", 0)

	rootFlagSet := flag.NewFlagSet("raiju", flag.ExitOnError)
	verbose := rootFlagSet.Bool("v", false, "increase log verbosity")

	btc2sat := &ffcli.Command{
		Name:       "btc2sat",
		ShortUsage: "raiju btc2sat <btc>",
		ShortHelp:  "Convert bitcoins to satoshis",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("btc2sat only takes one arg")
			}

			if *verbose {
				cmdLog.Printf("converting %s btc to sats", args[0])
			}

			btc, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("unable to parse arg: %s", args[0])
			}

			fmt.Fprintln(os.Stdout, raiju.Btc2sat(btc))
			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "raiju [flags] <subcommand>",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{btc2sat},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		// no need to output redundant message
		if err != flag.ErrHelp {
			cmdLog.Fatalln(err)
		} else {
			os.Exit(1)
		}
	}
}
