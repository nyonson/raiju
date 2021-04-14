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

# usage

All of Raiju's subcommands can be listed with the global help flag.

```
raiju -h
```

## btc2sats

Quick conversion from btc to the more popular satoshi granularity.

```
raiju btc2sats .000434

```

# install

Raiju can be built and installed locally with `make`. It requires `go` on the system to be compiled. Specify a `BINDIR` to override the default.

```
make BINDIR=~/.local/bin install
```

## configuration

All flags can be set through environment variables.