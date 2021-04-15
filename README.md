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

## btc2sat

Quick conversion from btc to the smaller satoshi unit which is more popular in the lightning network.

```
raiju btc2sat .000434
43400
```

## span



# installation

Binaries are available for download from [releases](https://github.com/nyonson/raiju/releases).

```
curl -OL https://github.com/nyonson/raiju/releases/download/$VERSION/raiju-$VERSION-linux-amd64.tar.gz
tar -xvzf raiju-$VERSION-linux-amd64.tar.gz
mv raiju ~/.local/bin
```

Alternatively, Raiju can also be built locally.

## build locally

Raiju can be built and installed locally with `make`. It requires `go` on the system to be compiled. Specify a `BINDIR` to override the default.

```
make BINDIR=~/.local/bin install
```

## configuration

All flags can be set through environment variables.