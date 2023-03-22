```
      :::::::::      :::     ::::::::::: ::::::::::: :::    ::: 
     :+:    :+:   :+: :+:       :+:         :+:     :+:    :+:  
    +:+    +:+  +:+   +:+      +:+         +:+     +:+    +:+   
   +#++:++#:  +#++:++#++:     +#+         +#+     +#+    +:+    
  +#+    +#+ +#+     +#+     +#+         +#+     +#+    +#+     
 #+#    #+# #+#     #+#     #+#     #+# #+#     #+#    #+#      
###    ### ###     ### ###########  #####       ########            
```
- [overview](#overview)
- [commands](#commands)
  - [candidates](#candidates)
  - [fees](#fees)
  - [rebalance](#rebalance)
  - [reaper](#reaper)
- [installation](#installation)
- [configuration](#configuration)
- [node](#node)

# overview

Your friendly bitcoin lightning network helper.

Raiju is a CLI app which sits on top of a lightning node. It currently only supports the [lnd](https://github.com/lightningnetwork/lnd) implementation. 

Raiju helps automate the channel life-cycle: creating, liquidity management, and closing. The `candidates` command helps open the most efficient new channels. The `fees` and `rebalance` commands automate passive and active liquidity management. And finally, the `reaper` command exposes inefficient channels to close in order to better allocate resources. 

# commands 

All of Raiju's commands can be listed with the global help flag.

```
raiju -h
```

## candidates

Find the best nodes to open a channel. `candidates` lists nodes by distance descending.

Theoretically these are desirable nodes to open channels to because they are well connected, but far (a.k.a. fees) away from the current node. The `Distant Neighbors` metric is the number of channels that node has with distant nodes from the root node.

```
$ raiju candidates
Pubkey                                                              Alias                             Distance  Distant Neighbors  Capacity    Channels  Updated
029ef8a775117ba63662a1d1d92b8a184bb1758ed1e12b0cdbb5e92672ef695b73  Carnivore                         4         8                  14932925    8         2021-04-21 23:17:36 -0700 PDT
0390b5d4492dc2f5318e5233ab2cebf6d48914881a33ef6a9c6bcdbb433ad986d0  LNBIG.com [lnd-01]                3         547                2568240344  547       2021-04-22 12:20:14 -0700 PDT
02c91d6aa51aa940608b497b6beebcb1aec05be3c47704b682b3889424679ca490  LNBIG.com [lnd-21]                3         372                2134427027  372       2021-04-22 08:32:48 -0700 PDT
02a04446caa81636d60d63b066f2814cbd3a6b5c258e3172cbdded7a16e2cfff4c  ln.bitstamp.net [Bitstamp]        3         366                1621569578  366       2021-04-22 12:43:28 -0700 PDT
...
```

The `assume` flag allows you to see the remaining set of nodes assuming channels were opened to a candidate. This can be used to find a set of nodes to open channels too in single batch transaction in order to minimize on onchain fees.

By default, only clearnet nodes are listed. TOR nodes tend to be unreliable due to the nature of TOR.

## fees

Set channel fees based on the channel's current liquidity.

The idea here is to encourage passive channel re-balancing through fees. If a channel has a too much local liquidity, fees are lowered in order to encourage relatively more outbound transactions. Visa versa for a channel with too little local liquidity.

The strategy for fee amounts is hardcoded (although I might try to add some more in the future) based on the standard fee ppm flag. 

```
$ raiju fees -standard-liquidity-fee-ppm 200
```

### systemd automation

Automatically update fees weekly with a systemd service and timer.

Example `fees.service`:

```
[Unit]
Description=Set fees of LND node

[Service]
User=lightning
Group=lightning
Environment=RAIJU_HOST=localhost:10009
Environment=RAIJU_MAC_PATH=/home/lightning/.lnd/data/chain/bitcoin/mainnet/admin.macaroon
Environment=RAIJU_TLS_PATH=/home/lightning/.lnd/tls.cert
Environment=RAIJU_STANDARD_LIQUIDITY_FEE_PPM=200
ExecStart=/usr/local/bin/raiju fees
```

Example `fees.timer`:

```
[Unit]
Description=Set fees daily at 3am

[Timer]
OnCalendar=*-*-* 03:00:00

[Install]
WantedBy=timers.target
```

## rebalance

Circular rebalance a channel or all channels that aren't doing so hot liquidity-wise.

Where the `fees` command attempts to balance channels passively, this is an *active* approach where liquidity is manually pushed. The cost of active rebalancing are the lightning payment fees. While this command could be used to push large amounts of liquidity, the default settings are intended to just prod things in the right direction. The maximum fee ppm setting uses the low liquidity fee setting by default, which theoretically, means that even if a rebalance is instantly canceled out by a large payment at least fees are re-coup'd.   

The command takes two arguments:
1. A percentage of the channel capacity to attempt to rebalance per circular payment (the "step").
2. The maximum percentage of the channel capacity to attempt to rebalance.

A smaller step percentage will increase the likely hood of a successful payment, but might also increase fees a bit if the payment collects a lot of `base_fee`s.

If output channel and last hop node flags are specified, than just those channels will be rebalanced. The following example is pushing 1% of the channel `754031881261074944`'s capacity to the channel with the `03963169ddfcc5cc6afaff7764fa20dc2e21e9ed8ef0ff0ccd18137d62ae2e01f4` node. A max fee of `2000` ppm will be paid. 

```
$ raiju rebalance -last-hop-pubkey 03963169ddfcc5cc6afaff7764fa20dc2e21e9ed8ef0ff0ccd18137d62ae2e01f4 -out-channel-id 754031881261074944 1 1
```

If no out channel and last hop pubkey are given, the command will roll through all channels with high liquidity (as defined by raiju) and attempt to push it through channels of low liquidity (as defined by raiju).

```
$ raiju rebalance 1 1 
```

Why is the out channel a channel ID while the last hop (a.k.a. in channel) a pubkey? This is due to the lightning Network's protocol allowing for [non-strict forwarding](https://github.com/lightning/bolts/blob/master/04-onion-routing.md#non-strict-forwarding). There might be some ways to specify an in channel, but I haven't put too much thought into it yet. 

### systemd automation

Example `rebalance.service`:

```
[Unit]
Description=Rebalance channels of LND node

[Service]
User=lightning
Group=lightning
Environment=RAIJU_HOST=localhost:10009
Environment=RAIJU_MAC_PATH=/home/lightning/.lnd/data/chain/bitcoin/mainnet/admin.macaroon
Environment=RAIJU_TLS_PATH=/home/lightning/.lnd/tls.cert
Environment=RAIJU_STANDARD_LIQUIDITY_FEE_PPM=200
ExecStart=/usr/local/bin/raiju rebalance 1
```

Example `rebalance.timer`:

```
[Unit]
Description=Rebalance channels daily with a wiggle so not run at the same time every day

[Timer]
OnCalendar=*-*-* 00:00:00
RandomizedDelaySec=1h

[Install]
WantedBy=timers.target
```

## reaper

Find channels which should be closed and re-allocated.

```
$ raiju reaper
Channel ID          Pubkey   Capacity (BTC)  
859008852420919297  fewsats  0.1             
826864630068281345  Pinky    0.02   
```

# installation

Raiju requires `go` on the system to be compiled. `go install` creates a `raiju` executable.

```
$ git clone https://git.sr.ht/~yonson/raiju
$ cd raiju
$ go install cmd/raiju/raiju.go
```

# configuration

All flags can be found with the help flag `-h`.

*Global* flags (not subcommand flags) can be set on the CLI, through environment variables, or with a configuration file. Flags overwrite environment variables which overwrite the configuration file values.

Environment variables have a `RAIJU_` prefix on the flag name. For example, the global flag `host` can be set with the `RAIJU_HOST` environment variable.

A configuration file can be provided with the `-config` flag or the default location (for Linux it's `~/.config/raiju/config`) can be used. The configuration file format is a flag per line, whitespace delimited.

```
host localhost:10009
```

# node

Are you here looking for a node to open a channel too? Well, may I offer Riaju's node! Could always use the inbound: [`02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8`](https://amboss.space/node/02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8)
