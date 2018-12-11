# Go Ethereum over lachesis

## Note

Repo is the [github.com/ethereum/go-ethereum](https://github.com/ethereum/go-ethereum) fork.
It should be in the same local path as original: `$GOPATH/src/github.com/ethereum/`.


## Aims

Ethereum over lachesis network and consensus.
Full ethereum stack and lachesis performance.


## Changes

* Rename `p2p.Server` to `p2p.p2pServer`;
* Create **p2p/interface.go** (*`p2p.ServerInterface`, `p2p.Server` struct, `p2p.p2pServer`'s additionals methods, `p2p.LachesisAdapter` interface*);
* Create **eth/lachesis_adapter.go** (*`p2p.LachesisAdapter` implementation*);
* Add `p2p.Config.LachesisAdapter LachesisAdapter`;
* Create `LachesisAddrFlag` in **cmd/utils/flags.go**:
    ```
        LachesisAddrFlag = cli.StringFlag{
		Name:  "lachesis",
		Usage: "lachesis-node address",
	}
	. . .
	func setListenAddress(ctx *cli.Context, cfg *p2p.Config) {
		. . .
		if ctx.GlobalIsSet(LachesisAddrFlag.Name) {
			cfg.LachesisAdapter = eth.NewLachesisAdapter(ctx.GlobalString(LachesisAddrFlag.Name), cfg.Logger)
		}
	}
	. . .
	func SetEthConfig(ctx *cli.Context, stack *node.Node, cfg *eth.Config) {
		. . .
		if ctx.GlobalIsSet(LachesisAddrFlag.Name) {
			cfg.Ethash.PowMode = ethash.ModeFake
		}
	}
    ```
* Append `utils.LachesisAddrFlag` to:
    - `nodeFlags` in **cmd/geth/main.go**;
    - `AppHelpFlagGroups` in **cmd/geth/usage.go**;
    - `app.Flags` in **cmd/swarm/main.go**;
* Make `node.Node` create `.server` according to `.serverConfig.LachesisAdapter` and use `p2p.Server.AddProtocols()` at `.Start()`:
    ```
	var running *p2p.Server
	if n.serverConfig.LachesisAdapter == nil {
		running = p2p.NewServer(n.serverConfig)
		n.log.Info("Starting peer-to-peer node", "instance", n.serverConfig.Name)
	} else {
		running = lachesis.NewServer(n.serverConfig)
		n.log.Info("Using lachesis node", "address", n.serverConfig.LachesisAdapter.Address())
	}
	. . .
	for _, service := range services {
		running.AddProtocols(service.Protocols()...)
	}
    ```
* Create **lachesis/** package;


## Demo

see `lachesis/demo/`


## TODO:

* make eth.lachesisAdapter clever;
* switch `lachesis/demo/Dockerfile.geth` from local to origin "github.com/Fantom-foundation/go-lachesis" when stable;
