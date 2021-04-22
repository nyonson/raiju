package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"strconv"

	"github.com/lightninglabs/lndclient"
	"github.com/nyonson/raiju"
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"
)

// version is set by build tools during linking
var version = "undefined"

func main() {
	cmdLog := log.New(os.Stderr, "raiju: ", 0)

	rootFlagSet := flag.NewFlagSet("raiju", flag.ExitOnError)
	verbose := rootFlagSet.Bool("verbose", false, "increase log verbosity")

	// lnd flags
	host := rootFlagSet.String("host", "localhost:10009", "lnd host and port")
	tlsPath := rootFlagSet.String("tlsPath", "", "lnd's tls cert path, defaults to lnd's default")
	macDir := rootFlagSet.String("macDir", "", "lnd's macaroons directory, defaults to lnd's default")
	network := rootFlagSet.String("network", "mainnet", "lightning network")

	btcToSatCmd := &ffcli.Command{
		Name:       "btc-to-sat",
		ShortUsage: "raiju btc-to-sat <btc>",
		ShortHelp:  "Convert bitcoins to satoshis",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("btc-to-sat only takes one arg")
			}

			if *verbose {
				cmdLog.Printf("converting %s btc to sats", args[0])
			}

			btc, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("unable to parse arg: %s", args[0])
			}

			raiju.PrintBtcToSat(btc)
			return nil
		},
	}

	nbdFlagSet := flag.NewFlagSet("span", flag.ExitOnError)
	minCapacity := nbdFlagSet.Int64("minCapacity", int64(10000000), "Minimum capacity of a node to display")
	minChannels := nbdFlagSet.Int("minChannels", 5, "Minimum channels of a node to display")
	minDistance := nbdFlagSet.Int("minDistance", 2, "Minimum distance of a node to display")
	pubkey := nbdFlagSet.String("pubkey", "", "Node to span out from, defaults to lnd's")

	nbdCmd := &ffcli.Command{
		Name:       "nodes-by-distance",
		ShortUsage: "raiju nodes-by-distance",
		ShortHelp:  "List network nodes by distance from node",
		FlagSet:    nbdFlagSet,
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("nodes-by-distance doesn't take any arguements")
			}

			client, err := lndclient.NewBasicClient(*host, *tlsPath, *macDir, *network)

			if err != nil {
				return err
			}

			err = raiju.NodesByDistance(client, *pubkey, *minCapacity, *minChannels, *minDistance)

			if err != nil {
				return err
			}

			return nil
		},
	}

	versionCmd := &ffcli.Command{
		Name:       "version",
		ShortUsage: "raiju version",
		ShortHelp:  "Version of raiju",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("version does not take any args")
			}

			fmt.Fprintln(os.Stdout, version)
			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "raiju [global flags] <subcommand> [subcommand flags] [subcommand args]",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{btcToSatCmd, nbdCmd, versionCmd},
		Options:     []ff.Option{ff.WithEnvVarPrefix("RAIJU"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.PlainParser)},
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
