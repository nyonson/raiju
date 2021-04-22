```
########     ###    ####       ## ##     ## 
##     ##   ## ##    ##        ## ##     ## 
##     ##  ##   ##   ##        ## ##     ## 
########  ##     ##  ##        ## ##     ## 
##   ##   #########  ##  ##    ## ##     ## 
##    ##  ##     ##  ##  ##    ## ##     ## 
##     ## ##     ## ####  ######   #######  
```

# overview

Raiju is your friendly bitcoin lightning network helper.

![build status](https://github.com/nyonson/raiju/actions/workflows/build.yml/badge.svg)

# usage

All of Raiju's subcommands can be listed with the global help flag.

```
raiju -h
```

## btc-to-sat

Quick conversion from btc to the smaller satoshi unit which is more popular in the lightning network.

```
raiju btc2sat .000434
43400
```

## nodes-by-distance

```
raiju nodes-by-distance
Pubkey                                                              Alias                             Distance  Capacity    Channels  Updated
0390b5d4492dc2f5318e5233ab2cebf6d48914881a33ef6a9c6bcdbb433ad986d0  LNBIG.com [lnd-01]                3         2568240344  547       2021-04-21 11:57:39 -0700 PDT
02c91d6aa51aa940608b497b6beebcb1aec05be3c47704b682b3889424679ca490  LNBIG.com [lnd-21]                3         2132269524  370       2021-04-21 11:10:46 -0700 PDT
039edc94987c8f3adc28dab455efc00dea876089a120f573bd0b03c40d9d3fb1e1  LNBIG.com [lnd-32]                3         1829244867  301       2021-04-20 23:12:11 -0700 PDT
```

# installation

Binaries are available for download from [releases](https://github.com/nyonson/raiju/releases).

```
curl -OL https://github.com/nyonson/raiju/releases/download/$VERSION/raiju-$VERSION-linux-amd64.tar.gz
tar -xvzf raiju-$VERSION-linux-amd64.tar.gz

# move the executable to the preferred bin directory on the PATH
mv raiju ~/.local/bin
```

Alternatively, Raiju can also be built locally.

## build locally

Raiju can be built and installed locally with `make`. It requires `go` on the system to be compiled. Specify a `BINDIR` to override the default.

```
make BINDIR=~/.local/bin install
```

## configuration

All flags can be listed with the `-h` flag.

```
# global flags
raiju -h

# subcommand flags
raiju nodes-by-distance -h
```

*Global* flags (not subcommand flags) can also be set through environment variables or a configuration file. Flags overwrite environment variables which overwrite the configuration file.

Environment variables have a `RAIJU_` prefix appended to the flag name. For example, the global `host` flag can be set with the `RAIJU_HOST` environment variable.

To use a configuration file, a path must be provided with the `-config` global flag. The configuration file format is a flag per line, space delimited.

```
host localhost:10009
```