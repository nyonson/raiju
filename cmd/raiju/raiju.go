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
	standardLiquidityFeePPM := rootFlagSet.Float64("standard-liquidity-fee-ppm", 200, "Default fee in PPM for standard liquidity channels which is shared by subcommands")

	candidatesFlagSet := flag.NewFlagSet("candidates", flag.ExitOnError)
	minCapacity := candidatesFlagSet.Int64("min-capacity", 1000000, "Minimum capacity of a node in satoshis")
	minChannels := candidatesFlagSet.Int64("min-channels", 5, "Minimum channels of a node")
	minDistance := candidatesFlagSet.Int64("min-distance", 2, "Minimum distance of a node")
	minNeighborDistance := candidatesFlagSet.Int64("min-neighbor-distance", 2, "Minimum distance of a neighbor node")
	pubkey := candidatesFlagSet.String("pubkey", "", "Node to span out from, defaults to the connected node")
	assume := candidatesFlagSet.String("assume", "", "Comma separated pubkeys to assume channels too")
	limit := candidatesFlagSet.Int64("limit", 100, "Number of results")
	clearnet := candidatesFlagSet.Bool("clearnet", true, "Filter tor-only nodes")

	// Bump up from the default of 30s to 5m since a lot of raiju's commands are long pulls of data
	rpcTimeout := time.Minute * 5

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
				RPCTimeout:         rpcTimeout,
			}
			services, err := lndclient.NewLndServices(cfg)

			if err != nil {
				return err
			}

			c := lightning.New(services.Client, services.Client, services.Router)
			r := raiju.New(c)

			// using FieldsFunc to handle empty string case correctly
			raw := strings.FieldsFunc(*assume, func(c rune) bool { return c == ',' })
			assume := make([]lightning.PubKey, len(raw))
			for i, a := range raw {
				assume[i] = lightning.PubKey(a)
			}

			request := raiju.CandidatesRequest{
				PubKey:              lightning.PubKey(*pubkey),
				MinCapacity:         lightning.Satoshi(*minCapacity),
				MinChannels:         *minChannels,
				MinDistance:         *minDistance,
				MinNeighborDistance: *minNeighborDistance,
				MinUpdated:          time.Now().Add(-2 * 24 * time.Hour),
				Assume:              assume,
				Limit:               *limit,
				Clearnet:            *clearnet,
			}

			_, err = r.Candidates(ctx, request)

			return err
		},
	}

	feesFlagSet := flag.NewFlagSet("fees", flag.ExitOnError)
	standardLiquidityFeePPMOverride := feesFlagSet.Float64("standard-liquidity-fee-ppm", 0, "Override the default standard liquidity fee ppm")

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
				RPCTimeout:         rpcTimeout,
			}
			services, err := lndclient.NewLndServices(cfg)

			if err != nil {
				return err
			}

			c := lightning.New(services.Client, services.Client, services.Router)
			r := raiju.New(c)

			// default to standard fee, override with flag
			standard := *standardLiquidityFeePPM
			if *standardLiquidityFeePPMOverride != 0 {
				standard = *standardLiquidityFeePPMOverride
			}

			return r.Fees(ctx, raiju.NewLiquidityFees(standard))
		},
	}

	rebalanceFlagSet := flag.NewFlagSet("rebalance", flag.ExitOnError)
	outChannelID := rebalanceFlagSet.Uint64("out-channel-id", 0, "Send out of channel ID")
	lastHopPubkey := rebalanceFlagSet.String("last-hop-pubkey", "", "Receive from node")
	maxFeePPM := rebalanceFlagSet.Float64("max-fee-ppm", 0, "Override the default of low liquidity fee ppm based on global standard flag")

	rebalanceCmd := &ffcli.Command{
		Name:       "rebalance",
		ShortUsage: "raiju rebalance <step-percent> <max-percent>",
		ShortHelp:  "Send circular payment(s) to actively rebalance channels",
		LongHelp:   "If the output and input flags are set, a rebalance is attempted (both must be set together). If not, channels are grouped into three coarse grained buckets: standard, high, and low. Standard channels will be ignored since their liquidity is good. High channels will attempt to push the percent of their capacity at a time to the low channels, stopping if their liquidity improves enough or if all channels have been tried.",
		FlagSet:    rebalanceFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 2 {
				return errors.New("rebalance takes two args")
			}

			// must be set together
			if (*lastHopPubkey != "" && *outChannelID == 0) || (*outChannelID != 0 && *lastHopPubkey == "") {
				return errors.New("out-channel-id and last-hop-pubkey must be set together")
			}

			stepPercent, err := strconv.ParseFloat(args[0], 64)
			if err != nil {
				return fmt.Errorf("unable to parse arg: %s", args[0])
			}

			maxPercent, err := strconv.ParseFloat(args[1], 64)
			if err != nil {
				return fmt.Errorf("unable to parse arg: %s", args[1])
			}

			cfg := &lndclient.LndServicesConfig{
				LndAddress:         *host,
				Network:            lndclient.Network(*network),
				CustomMacaroonPath: *macPath,
				TLSPath:            *tlsPath,
				RPCTimeout:         rpcTimeout,
			}
			services, err := lndclient.NewLndServices(cfg)
			if err != nil {
				return err
			}

			c := lightning.New(services.Client, services.Client, services.Router)
			r := raiju.New(c)

			fees := raiju.NewLiquidityFees(*standardLiquidityFeePPM)

			// default to low liquidity fee, override with flag
			maxFee := fees.Low()
			if *maxFeePPM != 0 {
				maxFee = lightning.FeePPM(*maxFeePPM)
			}

			if *lastHopPubkey != "" {
				_, err = r.Rebalance(ctx, lightning.ChannelID(*outChannelID), lightning.PubKey(*lastHopPubkey), stepPercent, maxPercent, maxFee)
			} else {
				err = r.RebalanceAll(ctx, stepPercent, maxPercent, maxFee)
			}

			return err
		},
	}

	reaperFlagSet := flag.NewFlagSet("reaper", flag.ExitOnError)

	reaperCmd := &ffcli.Command{
		Name:       "reaper",
		ShortUsage: "raiju reaper",
		ShortHelp:  "Find unproductive channels",
		LongHelp:   "",
		FlagSet:    reaperFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("reaper does not take any args")
			}

			cfg := &lndclient.LndServicesConfig{
				LndAddress:         *host,
				Network:            lndclient.Network(*network),
				CustomMacaroonPath: *macPath,
				TLSPath:            *tlsPath,
				RPCTimeout:         rpcTimeout,
			}
			services, err := lndclient.NewLndServices(cfg)

			if err != nil {
				return err
			}

			c := lightning.New(services.Client, services.Client, services.Router)
			r := raiju.New(c)

			_, err = r.Reaper(ctx)

			return err
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "raiju [global flags] <subcommand> [subcommand flags] [subcommand args]",
		FlagSet:     rootFlagSet,
		Subcommands: []*ffcli.Command{candidatesCmd, feesCmd, rebalanceCmd, reaperCmd},
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
