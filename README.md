```
########     ###    ####       ## ##     ## 
##     ##   ## ##    ##        ## ##     ## 
##     ##  ##   ##   ##        ## ##     ## 
########  ##     ##  ##        ## ##     ## 
##   ##   #########  ##  ##    ## ##     ## 
##    ##  ##     ##  ##  ##    ## ##     ## 
##     ## ##     ## ####  ######   #######  
```
![build status](https://github.com/nyonson/raiju/actions/workflows/build.yml/badge.svg)
- [overview](#overview)
- [usage](#usage)
  - [commands](#commands)
    - [btc-to-sat](#btc-to-sat)
    - [nodes-by-distance](#nodes-by-distance)
  - [patterns](#patterns)
- [installation](#installation)
  - [build locally](#build-locally)
- [configuration](#configuration)

# overview

Raiju is your friendly bitcoin lightning network helper.

# usage

Raiju is a CLI app which sits on top of a running lightning node instance. It only supports the [lnd](https://github.com/lightningnetwork/lnd) node implementation. Commands call out to the node and then perform anaylsis on the data returned.

## commands

All of Raiju's subcommands can be listed with the global help flag.

```
raiju -h
```

### btc-to-sat

Quick conversion from btc to the smaller satoshi unit.

```
raiju btc2sat .000434
43400
```

### nodes-by-distance

Lists nodes by distance and capacity descending. Theoretically these are desirable nodes to open channels to because they are well connected, but far (a.k.a. many fees) away from the current node.

```
raiju nodes-by-distance
Pubkey                                                              Alias                             Distance  Capacity    Channels  Updated
0390b5d4492dc2f5318e5233ab2cebf6d48914881a33ef6a9c6bcdbb433ad986d0  LNBIG.com [lnd-01]                3         2568240344  547       2021-04-21 11:57:39 -0700 PDT
02c91d6aa51aa940608b497b6beebcb1aec05be3c47704b682b3889424679ca490  LNBIG.com [lnd-21]                3         2132269524  370       2021-04-21 11:10:46 -0700 PDT
039edc94987c8f3adc28dab455efc00dea876089a120f573bd0b03c40d9d3fb1e1  LNBIG.com [lnd-32]                3         1829244867  301       2021-04-20 23:12:11 -0700 PDT
...
```

## patterns

Raiju is stateless, but running commands on a cron schedule could provide some insights over time.

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

All flags can be found with the `-h` flag.

```
# list global flags and subcommands
raiju -h

# list a subcommand's flags
raiju nodes-by-distance -h
```

*Global* flags (not subcommand flags) can also be set through environment variables or a configuration file. Flags overwrite environment variables which overwrite the configuration file values.

Environment variables have a `RAIJU_` prefix appended to the flag name. For example, the global flag `host` can be set with the `RAIJU_HOST` environment variable.

A configuration file can be provided with the `-config` flag or the default location (for Linux it's `~/.config/raiju/config`) can be used. The configuration file format is a flag per line, whitespace delimited.

```
host localhost:10009
```