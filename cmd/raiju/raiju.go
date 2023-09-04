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
	"github.com/rivo/tview"

	"github.com/nyonson/raiju"
	"github.com/nyonson/raiju/lightning"
	"github.com/nyonson/raiju/view"
)

const (
	// Bump up from the default of 30s to 5m since a lot of raiju's commands are long pulls of data
	rpcTimeout = time.Minute * 5
)

func parseFees(thresholds string, fees string, stickiness float64) (raiju.LiquidityFees, error) {
	// using FieldsFunc to handle empty string case correctly
	rawThresholds := strings.FieldsFunc(thresholds, func(c rune) bool { return c == ',' })
	tfs := make([]float64, len(rawThresholds))
	for i, t := range rawThresholds {
		tf, err := strconv.ParseFloat(t, 64)
		if err != nil {
			return raiju.LiquidityFees{}, err
		}
		tfs[i] = tf
	}

	rawFees := strings.FieldsFunc(fees, func(c rune) bool { return c == ',' })
	ffs := make([]lightning.FeePPM, len(rawFees))
	for i, f := range rawFees {
		ff, err := strconv.ParseFloat(f, 64)
		if err != nil {
			return raiju.LiquidityFees{}, err
		}
		ffs[i] = lightning.FeePPM(ff)
	}

	lf, err := raiju.NewLiquidityFees(tfs, ffs, stickiness)
	if err != nil {
		return raiju.LiquidityFees{}, err
	}

	return lf, nil
}

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
	// fees flags
	liquidityThresholds := rootFlagSet.String("liquidity-thresholds", "85,15", "Comma separated local liquidity percent thresholds")
	liquidityFees := rootFlagSet.String("liquidity-fees", "5,50,500", "Comma separated local liquidity-based fees PPM")
	liquidityStickiness := rootFlagSet.Float64("liquidity-stickiness", 0, "Percent of a channel capacity beyond threshold to wait before changing fees from settings attempting to improve liquidity")

	candidatesFlagSet := flag.NewFlagSet("candidates", flag.ExitOnError)
	minCapacity := candidatesFlagSet.Int64("min-capacity", 1000000, "Minimum capacity of a node in satoshis")
	minChannels := candidatesFlagSet.Int64("min-channels", 1, "Candidate must have at least this many channels")
	minDistance := candidatesFlagSet.Int64("min-distance", 2, "Candidate must be at least this far away (0 is root node and 1 is direct connection)")
	minDistantNeighbors := candidatesFlagSet.Int64("min-distant-neighbors", 0, "Candidate must have a minimum number of distant neighbors")
	pubkey := candidatesFlagSet.String("pubkey", "", "Node to span out from, defaults to the connected node")
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

			if *minDistance < 2 {
				return errors.New("min-distance must be greater than 1")
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
			defer services.Close()

			c := lightning.NewLndClient(services, *network)
			f, err := parseFees(*liquidityThresholds, *liquidityFees, *liquidityStickiness)
			if err != nil {
				return err
			}

			r := raiju.New(c, f)

			// using FieldsFunc to handle empty string case correctly
			a := strings.FieldsFunc(*assume, func(c rune) bool { return c == ',' })
			assume := make([]lightning.PubKey, len(a))
			for i, p := range a {
				assume[i] = lightning.PubKey(p)
			}

			request := raiju.CandidatesRequest{
				PubKey:              lightning.PubKey(*pubkey),
				MinCapacity:         lightning.Satoshi(*minCapacity),
				MinChannels:         *minChannels,
				MinDistance:         *minDistance,
				MinDistantNeighbors: *minDistantNeighbors,
				MinUpdated:          time.Now().Add(-2 * 24 * time.Hour),
				Assume:              assume,
				Limit:               *limit,
				Clearnet:            *clearnet,
			}

			cmdLog.Printf("filtering candidates by capacity: %d, channels: %d, distance: %d, distant neighbors: %d\n", request.MinCapacity, request.MinChannels, request.MinDistance, request.MinDistantNeighbors)

			candidates, err := r.Candidates(ctx, request)
			if err != nil {
				return err
			}

			view.TableNodes(candidates)

			return nil
		},
	}

	feesCmd := &ffcli.Command{
		Name:       "fees",
		ShortUsage: "raiju fees",
		ShortHelp:  "Set channel fees based on liquidity to passively rebalance channels",
		LongHelp:   "Channels are grouped depending on the local liquidity thresholds setting and have fees applied based on the local liquidity fees setting.",
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
			defer services.Close()

			c := lightning.NewLndClient(services, *network)
			f, err := parseFees(*liquidityThresholds, *liquidityFees, *liquidityStickiness)
			if err != nil {
				return err
			}

			view.TableFees(f)

			r := raiju.New(c, f)

			uc, ec, err := r.Fees(ctx)
			if err != nil {
				return err
			}

			// listen for first update which sets the fees, then exit
			select {
			case u := <-uc:
				for id, fee := range u {
					cmdLog.Printf("channel %d updated to %f fee PPM", id, fee)
				}
			case err := <-ec:
				return err
			}

			return nil
		},
	}

	rebalanceFlagSet := flag.NewFlagSet("rebalance", flag.ExitOnError)
	maxFeePPM := rebalanceFlagSet.Float64("max-fee-ppm", 0, "Override the default of low liquidity fee ppm based on global standard flag")

	rebalanceCmd := &ffcli.Command{
		Name:       "rebalance",
		ShortUsage: "raiju rebalance <max-percent>",
		ShortHelp:  "Send circular payment(s) to actively rebalance channels",
		LongHelp:   "Attempts to move liquidity from the channels with the highest local liquidity to the lowest.",
		FlagSet:    rebalanceFlagSet,
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 1 {
				return errors.New("rebalance takes one arg")
			}

			maxPercent, err := strconv.ParseFloat(args[0], 64)
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
			defer services.Close()

			c := lightning.NewLndClient(services, *network)
			f, err := parseFees(*liquidityThresholds, *liquidityFees, *liquidityStickiness)
			if err != nil {
				return err
			}

			view.TableFees(f)

			r := raiju.New(c, f)

			// default to low liquidity fee, override with flag
			maxFee := f.RebalanceFee()
			if *maxFeePPM != 0 {
				maxFee = lightning.FeePPM(*maxFeePPM)
			}

			cmdLog.Println("Rebalancing channels...")
			rebalanced, err := r.Rebalance(ctx, maxPercent, maxFee)
			if err != nil {
				return err
			}
			for id, percent := range rebalanced {
				cmdLog.Printf("rebalanced %f percent of channel %d\n", percent, id)
			}

			return nil
		},
	}

	daemonCmd := &ffcli.Command{
		Name:       "daemon",
		ShortUsage: "raiju daemon",
		ShortHelp:  "Daemon process running subcommands",
		LongHelp:   "Long running service process to passively manage fees and periodically active rebalances.",
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("fees does not take any args")
			}

			f, err := parseFees(*liquidityThresholds, *liquidityFees, *liquidityStickiness)
			if err != nil {
				return err
			}

			view.TableFees(f)

			// periodically rebalance, don't panic in this routine since its best effort
			go func() {
				cfg := &lndclient.LndServicesConfig{
					LndAddress:         *host,
					Network:            lndclient.Network(*network),
					CustomMacaroonPath: *macPath,
					TLSPath:            *tlsPath,
					RPCTimeout:         rpcTimeout,
				}
				services, err := lndclient.NewLndServices(cfg)
				if err != nil {
					cmdLog.Printf("Unable to connect to lnd in order to rebalance %s", err)
					return
				}
				defer services.Close()

				c := lightning.NewLndClient(services, *network)
				r := raiju.New(c, f)
				for range time.Tick(time.Duration(4) * time.Hour) {
					func() {
						cmdLog.Println("Rebalancing channels...")
						rebalanced, err := r.Rebalance(ctx, 5.0, f.RebalanceFee())
						if err != nil {
							cmdLog.Printf("Unable to to rebalance %s", err)
							return
						}
						for id, percent := range rebalanced {
							cmdLog.Printf("rebalanced %f percent of channel %d\n", percent, id)
						}
					}()
				}
			}()

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
			defer services.Close()

			c := lightning.NewLndClient(services, *network)
			r := raiju.New(c, f)

			uc, ec, err := r.Fees(ctx)
			if err != nil {
				return err
			}

			// listen for updates
			for {
				select {
				case u := <-uc:
					for id, fee := range u {
						cmdLog.Printf("channel %d updated to %f fee PPM", id, fee)
					}
				case err := <-ec:
					return err
				}
			}

			return nil
		},
	}

	root := &ffcli.Command{
		ShortUsage:  "raiju [global flags] [subcommand] [subcommand flags] [subcommand args]",
		FlagSet:     rootFlagSet,
		ShortHelp:   "Interactive dashboard",
		LongHelp:    "If given no subcommand, fire up an interactive dashboard that uses the subcommands under the hood.",
		Subcommands: []*ffcli.Command{candidatesCmd, daemonCmd, feesCmd, rebalanceCmd},
		Options:     []ff.Option{ff.WithEnvVarPrefix("RAIJU"), ff.WithConfigFileFlag("config"), ff.WithConfigFileParser(ff.PlainParser), ff.WithAllowMissingConfigFile(true)},
		Exec: func(ctx context.Context, args []string) error {
			if len(args) != 0 {
				return errors.New("raiju does not take any args")
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
			defer services.Close()

			c := lightning.NewLndClient(services, *network)
			f, err := parseFees(*liquidityThresholds, *liquidityFees, *liquidityStickiness)
			if err != nil {
				return err
			}

			r := raiju.New(c, f)

			app := tview.NewApplication().EnableMouse(true)
			// "column"
			flex := tview.NewFlex().SetDirection(0)
			flex.SetBorder(true).SetTitle("raiju")
			app.SetRoot(flex, true)

			viewChannels, err := view.ViewChannels(ctx, r)
			if err != nil {
				return err
			}

			viewCandidates, err := view.ViewCandidates(ctx, r)
			if err != nil {
				return err
			}

			flex.AddItem(viewChannels, 0, 3, true)
			flex.AddItem(viewCandidates, 0, 1, true)

			if err := app.Run(); err != nil {
				return err
			}

			return nil
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
