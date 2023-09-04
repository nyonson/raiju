```
      :::::::::      :::     ::::::::::: ::::::::::: :::    ::: 
     :+:    :+:   :+: :+:       :+:         :+:     :+:    :+:  
    +:+    +:+  +:+   +:+      +:+         +:+     +:+    +:+   
   +#++:++#:  +#++:++#++:     +#+         +#+     +#+    +:+    
  +#+    +#+ +#+     +#+     +#+         +#+     +#+    +#+     
 #+#    #+# #+#     #+#     #+#     #+# #+#     #+#    #+#      
###    ### ###     ### ###########  #####       ########            
```

**Your friendly bitcoin lightning network helper.**

`raiju` is a CLI app which sits on top of a lightning node and brings some smarts (perhaps that is debateable) to the channel life-cycle: open, manage, and close. `raiju` only supports the [lnd](https://github.com/lightningnetwork/lnd) node implementation at the moment.

- [subcommands](#subcommands)
  - [candidates](#candidates)
  - [fees](#fees)
  - [rebalance](#rebalance)
  - [daemon](#daemon)
- [installation](#installation)
- [configuration](#configuration)
- [node](#node)

# subcommands 

All of `raiju`'s subcommands can be listed with the global help flag, `raiju -h`, and each command has its own help (e.g. `raiju candidates -h`).

## candidates

**Open the most efficient channels**

List the best nodes to open a channel to from the current node. `candidates` does not automatically open any channels, that needs to be done out-of-band with a different tool such as `lncli`. `candidates` just lists suggestions and is not intended to be automated (for now...). 

The current node has distance `0` to itself and distance `1` to the nodes it has channels with. A node with distance `2` has a channel with a node the current node is connected too, but no channel with the current node, and so on. "Distant Neighbors" are distant (greater than `2`) from the current node, but have a channel with the candidate. By default, only nodes with clearnet addresses are listed. TOR-only nodes tend to be unreliable due to the nature of TOR.

```
$ raiju candidates
Pubkey                                                              Alias                             Distance  Distant Neighbors  Capacity    Channels  Updated
029ef8a775117ba63662a1d1d92b8a184bb1758ed1e12b0cdbb5e92672ef695b73  Carnivore                         4         8                  14932925    8         2021-04-21 23:17:36 -0700 PDT
0390b5d4492dc2f5318e5233ab2cebf6d48914881a33ef6a9c6bcdbb433ad986d0  LNBIG.com [lnd-01]                3         547                2568240344  547       2021-04-22 12:20:14 -0700 PDT
02c91d6aa51aa940608b497b6beebcb1aec05be3c47704b682b3889424679ca490  LNBIG.com [lnd-21]                3         372                2134427027  372       2021-04-22 08:32:48 -0700 PDT
02a04446caa81636d60d63b066f2814cbd3a6b5c258e3172cbdded7a16e2cfff4c  ln.bitstamp.net [Bitstamp]        3         366                1621569578  366       2021-04-22 12:43:28 -0700 PDT
...
```

The `assume` flag allows you to see the remaining candidates and updated stats assuming channels were opened to the given nodes. This can be used to find a set of nodes to open channels too in single batch transaction in order to minimize on onchain fees.

From a "make money routing" perspective, theoretically, these most distant nodes with the most distant neighbor connections are good to open a channel to for some off the beaten path efficient routing vs. just connecting to the biggest node in the network. Your node could offer cheaper, better routing between two "clusters" of nodes than the biggest nodes. From a "make the network stronger in general" perspective, the hope is that this strategy creates a more decentralized network vs. everything being dependent on a handful of large hub nodes. 

## fees

**Passively manage channel liquidity**

Set channel fees based on the channel's current liquidity. The idea here is to encourage passive channel re-balancing through fees. If a channel has a too much local liquidity (high), fees are lowered in order to encourage relatively more outbound transactions. Visa versa for a channel with too little local liquidity (low). 

The global `-liquidity-thresholds` flag determines how channels are grouped into liquidity buckets, while the `-liquidity-fees` flag determines the fee settings applied to those groups. For example, if thresholds are set to `80,20` and fees set to `5,50,500`, then channels with over 80% local liquidity will have a 5 PPM fee, channels between 80% and 20% local liquidity will have a 50 PPM fee, and channels with less than 20% liquidity will have a 500 PPM fee.

The `-liquidity-stickiness` attempts to avoid extra gossip by waiting for channels to return to a healthier liquidity state before changing fees. If using the same settings as before, plus a stickiness setting of 5%, if a channel moves from 19% liquidity to 23% liquidity it will still have a 500 PPM fee. It needs to move to something better than 25% (20% + 5%) before the fee will change. The stickiness setting only applies to liquidity moving in a healthy (towards center) direction. If you are drastically changing your fee settings, you probably want to set stickiness to 0 temporarily to ensure fees are updated.

The `-liquidity-thresholds`, `-liquidity-fees`, and `-liquidity-stickiness` are global (not `fees` specific) because they are also used in the `rebalance` command to help coordinate the right amount of fees to pay in active rebalancing.

`fees` follows the [zero-base-fee movement](http://www.rene-pickhardt.de/). I am honestly not sure if this is financially sound, but I appreciate the simpler mental model of only thinking in ppm.

## rebalance

**Actively manage channel liquidity**

Circular rebalance channels that aren't doing so hot liquidity-wise.

Where the `fees` command attempts to balance channels passively, this is an *active* approach where liquidity is manually pushed. The cost of active rebalancing are the lightning payment fees. While this command could be used to push large amounts of liquidity, the default settings are intended to just prod things in the right direction. 

The maximum fee ppm setting defaults to the low liquidity fee setting used by the `fees` command. Theoretically, this means that even if a rebalance is instantly canceled out by a large payment at least fees are re-coup'd.

The command takes one argument, the maximum percentage of the channel capacity to attempt to rebalance.

The command will roll through channels with high liquidity and attempt to push it through channels of low liquidity. High and low are defined by the defined by the global `-liquidity-thresholds` flag. For example, if liquidity thresholds is set to `80,20`, channels with local liquidity over 80% are considered "high" and channels with local liquidity under 20% are considered "low".

## daemon

The `daemon` subcommand keeps the process alive listening for channel updates that trigger fee updates (e.g. a channel's liquidity sinks below the low level and needs its fees updated). It also periodically calls `rebalance` under the hood to *actively* balance channel liquidity.

### systemd automation

Example `fees.service`:

```
[Unit]
Description=Monitor LND node
Wants=lnd.service
After=lnd.service

[Service]
User=lightning
Group=lightning
Restart=always
Environment=RAIJU_HOST=localhost:10009
Environment=RAIJU_MAC_PATH=/home/lightning/.lnd/data/chain/bitcoin/mainnet/admin.macaroon
Environment=RAIJU_TLS_PATH=/home/lightning/.lnd/tls.cert
Environment=RAIJU_LIQUIDITY_FEES=5,50,500
Environment=RAIJU_LIQUIDITY_STICKINESS=5
ExecStart=/usr/local/bin/raiju daemon

[Install]
WantedBy=multi-user.target
```

# installation

To install from source, `raiju` requires `go` on the system. `go install` creates a `raiju` executable.

```
$ go install github.com/nyonson/raiju/cmd/raiju@latest
```

If a container is preferred, `raiju` images are published at `ghcr.io/nyonson/raiju`. 

```
docker pull ghcr.io/nyonson/raiju:v0.7.1
```

A little more configuration is required to pass along settings to the container.

```
docker run -it \
  -v /admin.macaroon:/admin.macaroon:ro -v /tls.cert:/tls.cert:ro \
  ghcr.io/nyonson/raiju:v0.7.1 \
  -host 192.168.1.187:10009 -mac-path admin.macaroon -tls-path tls.cert
  candidates
```

* Ensure the tls certificate and macaroon are mounted in the container, in the above example they are both mounted to the root of the container's filesystem and then their paths are passed in as cli flags.
* The container may need to be attached to a network depending on your network. 

# configuration

*Global* flags (can be found with `raiju -h`) can be set through environment variables or with a configuration file. CLI flags overwrite environment variables which overwrite the configuration file values.

Environment variables have a `RAIJU_` prefix on the flag name. For example, the global flag `host` can be set with the `RAIJU_HOST` environment variable.

A configuration file can be provided with the `-config` flag or the default location (for Linux it's `~/.config/raiju/config`) can be used. The configuration file format is a flag per line, whitespace delimited.

```
host localhost:10009
```

# node

Are you here looking for a node to open a channel too? Well, may I offer `raiju`'s node! Could always use the inbound: [`02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8`](https://amboss.space/node/02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8)
