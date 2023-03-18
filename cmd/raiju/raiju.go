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
	"github.com/peterbourgon/ff/v3"
	"github.com/peterbourgon/ff/v3/ffcli"

	"github.com/nyonson/raiju"
	"github.com/nyonson/raiju/lightning"
)

func main() {
	cmdLog := log.New(os.Stderr, "raiju: ", 0)

	rootFlagSet := flag.NewFlagSet("raiju", flag.ExitOnError)

	// hooked up to ff with WithConfigFileFlag
	var defaultConfigFile string
	if d, err := os.UserConfigDir(); err == nil {
		defaultConfigFile = filepath.Join(d, "raiju", "config")
	}
	rootFlagSet.String("config", defaultConfigFile, "configuration file path")

	// lnd flags
	host := rootFlagSet.String("host", "localhost:10009", "LND host with port")
	tlsPath := rootFlagSet.String("tls-path", "", "LND node tls certificate")
	macPath := rootFlagSet.String("mac-path", "", "Macaroon with necessary permissions for lnd node")
	network := rootFlagSet.String("network", "mainnet", "The bitcoin network")
	// liquidity flags
	highFeePPM := rootFlagSet.Float64("high-fee-ppm", 2000, "Default high liquidity setting shared by commands")

	candidatesFlagSet := flag.NewFlagSet("candidates", flag.ExitOnError)
	minCapacity := candidatesFlagSet.Int64("min-capacity", 10000000, "Minimum capacity of a node")
	minChannels := candidatesFlagSet.Int64("min-channels", 5, "Minimum channels of a node")
	minDistance := candidatesFlagSet.Int64("min-distance", 2, "Minimum distance of a node")
	minNeighborDistance := candidatesFlagSet.Int64("min-neighbor-distance", 2, "Minimum distance of a neighbor node")
	pubkey := candidatesFlagSet.String("pubkey", "", "Node to span out from, defaults to lnd's")
	assume := candidatesFlagSet.String("assume", "", "Comma separated pubkeys to assume channels too")
	limit := candidatesFlagSet.Int64("limit", 100, "Number of results")
	clearnet := candidatesFlagSet.Bool("clearnet", true, "Filter tor-only nodes")

	candidatesCmd := &ffcli.Command{
		Name:       "candidates",
		ShortUsage: "raiju candidates",
		ShortHelp:  "List candidate nodes by distance from node and centralization",
		LongHelp:   "Nodes are listed in descending order based on a few calculated metrics. The dominant metric is distance from the root node. Next is 'distant neighbors' which is the number of direct neighbors a node has that are distant from the root node.",
		FlagSet:    candidatesFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("candidates doesn't take any arguments")
			}

			cfg := &lndclient.LndServicesConfig{
				LndAddress:         *host,
				Network:            lndclient.Network(*network),
				CustomMacaroonPath: *macPath,
				TLSPath:            *tlsPath,
			}
			services, err := lndclient.NewLndServices(cfg)

			if err != nil {
				return err
			}

			c := lightning.New(services.Client, services.Client, services.Router)
			r := raiju.New(c)

			request := raiju.CandidatesRequest{
				Pubkey:              *pubkey,
				MinCapacity:         *minCapacity,
				MinChannels:         *minChannels,
				MinDistance:         *minDistance,
				MinNeighborDistance: *minNeighborDistance,
				MinUpdated:          time.Now().Add(-2 * 24 * time.Hour),
				// using FieldsFunc to handle empty string case correctly
				Assume:   strings.FieldsFunc(*assume, func(c rune) bool { return c == ',' }),
				Limit:    *limit,
				Clearnet: *clearnet,
			}

			nodes, err := r.Candidates(ctx, request)

			if err != nil {
				return err
			}

			raiju.PrintNodes(nodes)

			return nil
		},
	}

	feesFlagSet := flag.NewFlagSet("fees", flag.ExitOnError)
	highFeePPMiOverride := feesFlagSet.Float64("high-fee-ppm", 0, "Override the default high fee ppm")

	feesCmd := &ffcli.Command{
		Name:       "fees",
		ShortUsage: "raiju fees",
		ShortHelp:  "Set channel fees based on liquidity to passively rebalance channels",
		LongHelp:   "Channels are grouped into three coarse grained buckets: standard, high, and low. Channels with standard liquidity will have the standard fee applied. Channels with high liquidity will have a 10x the standard fee applied to discourage routing. And channels with low liquidity will have 1/10 the standard fee applied to encourage routing.",
		FlagSet:    feesFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("fees does not take any args")
			}

			cfg := &lndclient.LndServicesConfig{
				LndAddress:         *host,
				Network:            lndclient.Network(*network),
				CustomMacaroonPath: *macPath,
				TLSPath:            *tlsPath,
			}
			services, err := lndclient.NewLndServices(cfg)

			if err != nil {
				return err
			}

			c := lightning.New(services.Client, services.Client, services.Router)
			r := raiju.New(c)

			// default to high fee, override with flag
			high := *highFeePPM
			if *highFeePPMiOverride != 0 {
				high = *highFeePPMiOverride
			}

			r.Fees(ctx, lightning.FeePPM(high))

			return nil
		},
	}

	rebalanceFlagSet := flag.NewFlagSet("rebalance", flag.ExitOnError)
	outChannelID := rebalanceFlagSet.Uint64("out-channel-id", 0, "Send out of channel ID")
	lastHopPubkey := rebalanceFlagSet.String("last-hop-pubkey", "", "Receive from node")
	maxFeePPM := rebalanceFlagSet.Float64("max-fee-ppm", 0, "Override the default high fee ppm")

	rebalanceCmd := &ffcli.Command{
		Name:       "rebalance",
		ShortUsage: "raiju rebalance <percent>",
		ShortHelp:  "Send circular payment(s) to actively rebalance channels",
		LongHelp:   "If the output and input flags are set, a rebalance is attempted (both must be set together). If not, channels are grouped into three coarse grained buckets: standard, high, and low. Standard channels will be ignored since their liquidity is good. High channels will attempt to push the percent of their capacity in liquidity at a time to the low channels, stopping if their liquidity improves enough or if all channels have been tried.",
		FlagSet:    rebalanceFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("rebalance takes one arg")
			}

			// must be set together
			if (*lastHopPubkey != "" && *outChannelID == 0) || (*outChannelID != 0 && *lastHopPubkey == "") {
				return errors.New("out-channel-id and last-hop-pubkey must be set together")
			}

			percent, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("unable to parse arg: %s", args[0])
			}

			cfg := &lndclient.LndServicesConfig{
				LndAddress:         *host,
				Network:            lndclient.Network(*network),
				CustomMacaroonPath: *macPath,
				TLSPath:            *tlsPath,
			}
			services, err := lndclient.NewLndServices(cfg)
			if err != nil {
				return err
			}

			c := lightning.New(services.Client, services.Client, services.Router)
			r := raiju.New(c)

			// default to high fee, override with flag
			max := *highFeePPM
			if *maxFeePPM != 0 {
				max = *maxFeePPM
			}

			if *lastHopPubkey != "" {
				err = r.Rebalance(ctx, *outChannelID, *lastHopPubkey, percent, lightning.FeePPM(max))
			} else {
				err = r.RebalanceAll(ctx, percent, lightning.FeePPM(max))
			}

			return err
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "raiju [global flags] <subcommand> [subcommand flags] [subcommand args]",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{candidatesCmd, feesCmd, rebalanceCmd},
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
