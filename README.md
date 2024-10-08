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

`raiju` sits on top of a lightning node and brings some smarts (perhaps that is debateable) to the channel opening and liquidity managment. `raiju` only supports the [lnd](https://github.com/lightningnetwork/lnd) node implementation at the moment.

- [commands](#commands)
  - [candidates](#candidates)
  - [fees](#fees)
  - [rebalance](#rebalance)
  - [daemon](#daemon)
- [installation](#installation)
  - [source](#source)
  - [container](#container)
  - [nix](#nix)
- [configuration](#configuration)
- [node](#node)

# commands 

All of `raiju`'s commands can be listed with the global help flag, `raiju -h`, and each command has its own help (e.g. `raiju candidates -h`).

## candidates

**Open the most efficient channels**

List the best nodes to open a channel to from the current node. `candidates` does not automatically open any channels, that needs to be done out-of-band with a different tool such as `lncli`. `candidates` just lists suggestions and is not intended to be automated (for now...). 

The current node has distance `0` to itself and distance `1` to the nodes it has channels with. A node with distance `2` has a channel with a node the current node is connected to, but no channel with the current node, and so on. "Distant Neighbors" are distant (greater than `2`) from the current node, but have a channel with the candidate. By default, only nodes with clearnet addresses are listed. TOR-only nodes tend to be unreliable due to the nature of TOR.

```
$ raiju candidates
Pubkey                                                              Alias                             Distance  Distant Neighbors  Capacity    Channels  Updated
029ef8a775117ba63662a1d1d92b8a184bb1758ed1e12b0cdbb5e92672ef695b73  Carnivore                         4         8                  14932925    8         2021-04-21 23:17:36 -0700 PDT
0390b5d4492dc2f5318e5233ab2cebf6d48914881a33ef6a9c6bcdbb433ad986d0  LNBIG.com [lnd-01]                3         547                2568240344  547       2021-04-22 12:20:14 -0700 PDT
02c91d6aa51aa940608b497b6beebcb1aec05be3c47704b682b3889424679ca490  LNBIG.com [lnd-21]                3         372                2134427027  372       2021-04-22 08:32:48 -0700 PDT
02a04446caa81636d60d63b066f2814cbd3a6b5c258e3172cbdded7a16e2cfff4c  ln.bitstamp.net [Bitstamp]        3         366                1621569578  366       2021-04-22 12:43:28 -0700 PDT
...
```

The `assume` flag allows you to see the remaining candidates and updated stats assuming channels were opened to the given nodes. This can be used to find a set of nodes to open channels to in a single batch transaction in order to minimize on onchain fees.

From a "make money routing" perspective, theoretically, these most distant nodes with the most distant neighbor connections are good to open a channel to for some off the beaten path efficient routing vs. just connecting to the biggest node in the network. Your node could offer cheaper, better routing between two "clusters" of nodes than the biggest nodes. From a "make the network stronger in general" perspective, the hope is that this strategy creates a more decentralized network vs. everything being dependent on a handful of large hub nodes. 

## fees

**Passively manage channel liquidity**

Set channel fees based on the channel's current liquidity. The idea here is to encourage passive channel rebalancing through fees. If a channel has a too much local liquidity, fees are lowered in order to encourage relatively more outbound transactions. _vice versa_ for a channel with too little local liquidity.

There are three liquidity flags which control when and what fees are applied to channels. 

* `Liquidity Fees` -- The feerates (in PPM) raiju applies to channels. The first option is applied to channels with too much local liquidity. The second is for balanced channels (in between the two threshold values). The third option is applied to channels with too little local liquidity
* `Liquidity Thresholds` -- Determines how channels are grouped into liquidity buckets, while the `-liquidity-fees` flag determines the fee settings applied to those groups. For example, if thresholds are set to `80,20` and fees set to `5,50,500`, then channels with over 80% local liquidity will have a 5 PPM fee, channels between 80% and 20% local liquidity will have a 50 PPM fee, and channels with less than 20% liquidity will have a 500 PPM fee.
* `Liquidity Stickiness` -- Attempts to avoid extra gossip by waiting for channels to return to a healthier liquidity state before changing fees. If using the same settings as before, plus a stickiness setting of 5%, if a channel moves from 19% liquidity to 23% liquidity it will still have a 500 PPM fee. It needs to move to something better than 25% (20% + 5%) before the fee will change. The stickiness setting only applies to liquidity moving in a healthy (towards center) direction. If you are drastically changing your fee settings, you probably want to set stickiness to 0 temporarily to ensure fees are updated.

The fees, thresholds, and stickiness are global settings (not specific to just the `fees` command) because they are also used in the `rebalance` command to help coordinate the right amount of fees to pay in active rebalancing.

`fees` follows the [zero-base-fee movement](http://www.rene-pickhardt.de/). I am honestly not sure if this is financially sound, but I appreciate the simpler mental model of only thinking in ppm.

`fees` also automatically applies some [flow control](https://blog.bitmex.com/the-power-of-htlc_maximum_msat-as-a-control-valve-for-better-flow-control-improved-reliability-and-lower-expected-payment-failure-rates-on-the-lightning-network/) to channels in order to encourage more rebalancing.

## rebalance

**Actively manage channel liquidity**

Circular rebalance channels that aren't doing so hot liquidity-wise.

Where the `fees` command attempts to balance channels passively, this is an *active* approach where liquidity is manually pushed. The cost of active rebalancing are the lightning payment fees. While this command could be used to push large amounts of liquidity, the default settings are intended to just prod things in the right direction. 

The maximum fee ppm setting defaults to the low liquidity fee setting used by the `fees` command. Theoretically, this means that even if a rebalance is instantly canceled out by a large payment at least fees are re-coup'd.

The command takes one argument, the maximum percentage of the channel capacity to attempt to rebalance.

The command will roll through channels with high liquidity and attempt to push it through channels of low liquidity. High and low are defined by the defined by the global liqudidity thresholds setting. For example, if liquidity thresholds is set to `80,20`, channels with local liquidity over 80% are considered "high" and channels with local liquidity under 20% are considered "low".

## daemon

This is where the magic really happens. The `daemon` command keeps the raiju process alive in order to listen for channel updates from LND when liquidity has shifted (e.g. a routed payment). As liquidity ebbs and flows, raiju instantly updates fees to *passively* push thigns in the right direction (e.g. a channel's liquidity sinks below the low level and needs its fees updated). The daemon process also periodically (every 12 hours) calls `rebalance` in order to *actively* balance liquidity to help move thigs along.

### systemd automation

Here is an example `raiju.service` systemd unit.

```
[Unit]
Description=Raiju
Requires=lnd.service
After=lnd.service

[Service]
Restart=always
Environment=RAIJU_HOST=localhost:10009
Environment=RAIJU_MAC_PATH=/path/to/lnd/admin.macaroon
Environment=RAIJU_TLS_PATH=/path/to/lnd/tls.cert
Environment=RAIJU_LIQUIDITY_FEES=5,50,500
Environment=RAIJU_LIQUIDITY_STICKINESS=5
Environment=RAIJU_LIQUIDITY_THRESHOLDS=80,20
ExecStart=/path/to/raiju daemon

[Install]
WantedBy=multi-user.target
```

# installation

Raiju can be installed from source, but a container and nix flake are also provided.

## source

`raiju` requires `go` on the system. `go install` creates a `raiju` executable.

```
$ go install github.com/nyonson/raiju/cmd/raiju@latest
```

## container

`raiju` images are published at `ghcr.io/nyonson/raiju`. 

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
* The container may need to be attached to a network depending on your system.

## nix

The nix flake sets up a developer shell and also builds an executable.

If running NixOS, the flake can be added and a systemd service easily setup to run the raiju daemon.

```
  inputs = {
    ...
    raiju.url = "github:nyonson/raiju";
    ...
  };

  outputs = { self, nixpkgs, raiju, ... }@inputs: {
      ...
      modules = [
       ...
        # Add Raiju and enable the daemon.
        raiju.nixosModules.default
        {
          services.raiju = {
            enable = true;
            rpcHost = "localhost:10009";
            macaroonFile = "/path/to/lnd/admin.macaroon";
            tlsCertificateFile = "/path/to/lnd/tls.cert";
            liquidityFees = [1 50 2500];
            liquidityThresholds = [80 20];
            liquidityStickiness = 10;
          };
        };
        ...
      ];
    };
  };
```

# configuration

*Global* flags (can be found with `raiju -h`) can be set through environment variables or with a configuration file. CLI flags overwrite environment variables which overwrite the configuration file values.

Environment variables have a `RAIJU_` prefix on the flag name. For example, the global flag `host` can be set with the `RAIJU_HOST` environment variable.

A configuration file can be provided with the `-config` flag or the default location (for Linux it's `~/.config/raiju/config`) can be used. The configuration file format is a flag per line, whitespace delimited.

```
host localhost:10009
```

# node

Are you here looking for a node to open a channel to? Well, may I offer `raiju`'s node! Could always use the inbound: [`02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8`](https://amboss.space/node/02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8)
