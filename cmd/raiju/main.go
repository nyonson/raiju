package main

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

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

	// hooked up to ff with WithConfigFileFlag
	var defaultConfigFile string
	if d, err := os.UserConfigDir(); err == nil {
		defaultConfigFile = filepath.Join(d, "raiju", "config")
	}
	rootFlagSet.String("config", defaultConfigFile, "configuration file path")

	// lnd flags
	host := rootFlagSet.String("host", "localhost:10009", "lnd host and port")
	tlsPath := rootFlagSet.String("tlsPath", "", "lnd's tls cert path, defaults to lnd's default")
	macDir := rootFlagSet.String("macDir", "", "lnd's macaroons directory, defaults to lnd's default")
	network := rootFlagSet.String("network", "mainnet", "lightning network")

	satsCmd := &ffcli.Command{
		Name:       "sats",
		ShortUsage: "raiju sats <btc>",
		ShortHelp:  "Convert bitcoins to satoshis",
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("sats only takes one arg")
			}

			btc, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("unable to parse arg: %s", args[0])
			}

			raiju.PrintBtcToSat(btc)
			return nil
		},
	}

	candidatesFlagSet := flag.NewFlagSet("span", flag.ExitOnError)
	minCapacity := candidatesFlagSet.Int64("minCapacity", int64(10000000), "Minimum capacity of a node")
	minChannels := candidatesFlagSet.Int("minChannels", 5, "Minimum channels of a node")
	minDistance := candidatesFlagSet.Int("minDistance", 2, "Minimum distance of a node")
	minNeighborDistance := candidatesFlagSet.Int("minNeighborDistance", 2, "Minimum distance of a neighbor node")
	pubkey := candidatesFlagSet.String("pubkey", "", "Node to span out from, defaults to lnd's")
	assume := candidatesFlagSet.String("assume", "", "Comma separated pubkeys to assume channels too")
	limit := candidatesFlagSet.Int("limit", 100, "Number of results")

	candidatesCmd := &ffcli.Command{
		Name:       "candidates",
		ShortUsage: "raiju candidates",
		ShortHelp:  "List candidate nodes by distance from node and centralization",
		LongHelp:   "Nodes are listed in decending order based on a few calculated metrics. The dominant metric is distance from the root node. Next is 'distant neighbors' which is the number of direct neighbors a node has that are distant from the root node.",
		FlagSet:    candidatesFlagSet,
		Exec: func(_ context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("candidates doesn't take any arguements")
			}

			client, err := lndclient.NewBasicClient(*host, *tlsPath, *macDir, *network)

			if err != nil {
				return err
			}

			app := raiju.App{Infoer: client, Grapher: client, Log: cmdLog, Verbose: *verbose}
			request := raiju.CandidatesRequest{
				Pubkey:              *pubkey,
				MinCapacity:         *minCapacity,
				MinChannels:         *minChannels,
				MinDistance:         *minDistance,
				MinNeighborDistance: *minNeighborDistance,
				MinUpdated:          time.Now().Add(-2 * 24 * time.Hour),
				// using FieldsFunc to handle empty string case correctly
				Assume: strings.FieldsFunc(*assume, func(c rune) bool { return c == ',' }),
				Limit:  *limit,
			}

			err = raiju.PrintCandidates(app, request)

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
		Subcommands: []*ffcli.Command{satsCmd, candidatesCmd, versionCmd},
		Options:     []ff.Option{ff.WithEnvVarPrefix("RAIJU"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.PlainParser), ff.WithAllowMissingConfigFile(true)},
		Exec: func(context.Context, []string) error {
			return flag.ErrHelp
		},
	}

	if err := root.ParseAndRun(context.Background(), os.Args[1:]); err != nil {
		// no need to output redundant message, just exit
		if err == flag.ErrHelp {
			os.Exit(1)
		}

		cmdLog.Fatalln(err)
	}
}
