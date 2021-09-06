```
      :::::::::      :::     ::::::::::: ::::::::::: :::    ::: 
     :+:    :+:   :+: :+:       :+:         :+:     :+:    :+:  
    +:+    +:+  +:+   +:+      +:+         +:+     +:+    +:+   
   +#++:++#:  +#++:++#++:     +#+         +#+     +#+    +:+    
  +#+    +#+ +#+     +#+     +#+         +#+     +#+    +#+     
 #+#    #+# #+#     #+#     #+#     #+# #+#     #+#    #+#      
###    ### ###     ### ###########  #####       ########            
```
![build status](https://github.com/nyonson/raiju/actions/workflows/build.yml/badge.svg)
- [overview](#overview)
- [usage](#usage)
  - [commands](#commands)
    - [sats](#sats)
    - [candidates](#candidates)
- [installation](#installation)
  - [build locally](#build-locally)
- [configuration](#configuration)
- [node](#node)

# overview

Raiju is your friendly bitcoin lightning network helper.

# usage

Raiju is a CLI app which sits on top of a running lightning node instance. It only supports the [lnd](https://github.com/lightningnetwork/lnd) node implementation. Raiju calls out to the node for network information and then performs anaylsis for insights.

## commands

All of Raiju's subcommands can be listed with the global help flag.

```
raiju -h
```

### sats

Quick conversion from btc to the smaller satoshi unit. We talk in sats.

```
raiju sats .000434
43400
```

### candidates

Lists nodes by distance descending. Theoretically these are desirable nodes to open channels to because they are well connected, but far (a.k.a. many fees) away from the current node. The `Distant Neighbors` metric is the number of channels that node has with distant nodes from the root node.

```
raiju candidates
Pubkey                                                              Alias                             Distance  Distant Neighbors  Capacity    Channels  Updated
029ef8a775117ba63662a1d1d92b8a184bb1758ed1e12b0cdbb5e92672ef695b73  Carnivore                         4         8                  14932925    8         2021-04-21 23:17:36 -0700 PDT
0390b5d4492dc2f5318e5233ab2cebf6d48914881a33ef6a9c6bcdbb433ad986d0  LNBIG.com [lnd-01]                3         547                2568240344  547       2021-04-22 12:20:14 -0700 PDT
02c91d6aa51aa940608b497b6beebcb1aec05be3c47704b682b3889424679ca490  LNBIG.com [lnd-21]                3         372                2134427027  372       2021-04-22 08:32:48 -0700 PDT
02a04446caa81636d60d63b066f2814cbd3a6b5c258e3172cbdded7a16e2cfff4c  ln.bitstamp.net [Bitstamp]        3         366                1621569578  366       2021-04-22 12:43:28 -0700 PDT
...
```

The `assume`` flag allows you to see the remaining set of nodes assuming channels were opened to a candidate. This can be used to find a set of nodes to open channels too in single batch transaction. Batch transactions minimize on onchain fees. 

# installation

Binaries are available for download from [releases](https://github.com/nyonson/raiju/releases).

```
# grab the binary
curl -OL https://github.com/nyonson/raiju/releases/download/$VERSION/raiju-$VERSION-linux-amd64.tar.gz

# optionally check md5 hash against the releases listed hash to ensure the correct binary
md5sum raiju-$VERSION-linux-amd64.tar.gz

# unpack the tarball
tar -xvzf raiju-$VERSION-linux-amd64.tar.gz

# move the executable to the preferred bin directory on the PATH
mv raiju ~/.local/bin
```

Alternatively, Raiju can also be built locally.

## build locally

Raiju can be built and installed locally with `make`. It requires `go` on the system to be compiled. Specify a `BINDIR` to override the default directory where `make` installs the executable.

```
git clone https://github.com/nyonson/raiju.git
cd raiju
make BINDIR=~/.local/bin install
```

# configuration

All flags can be found with the help flag `-h`.

```
# list global flags and subcommands
raiju -h

# list a subcommand's flags
raiju nodes-by-distance -h
```

*Global* flags (not subcommand flags) can be set on the CLI, through environment variables, or with a configuration file. Flags overwrite environment variables which overwrite the configuration file values.

Environment variables have a `RAIJU_` prefix appended to the flag name. For example, the global flag `host` can be set with the `RAIJU_HOST` environment variable.

A configuration file can be provided with the `-config` flag or the default location (for Linux it's `~/.config/raiju/config`) can be used. The configuration file format is a flag per line, whitespace delimited.

```
host localhost:10009
```

# node

Are you here looking for a node to open a channel too? Well, may I offer Riaju's node! Could always use the inbound: [`02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8`](https://1ml.com/node/02b6867b56ca1b6a4548b97b009152683fa366bfa1b14119c8f9992e1acacbe1c8)