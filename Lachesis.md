# Go Ethereum over lachesis

## Note

Repo is the [github.com/ethereum/go-ethereum](https://github.com/ethereum/go-ethereum) fork.
It should be in the same local path as original: `$GOPATH/src/github.com/ethereum/`.


## Aim

Ethereum over lachesis network and consensus.
Full ethereum stack and lachesis performance.


## Changes

* Rename p2p.Server to p2p.p2pServer;
* Create p2p/interface.go: p2p.ServerInterface, p2p.Server struct, p2p.p2pServer's additionals methods;
* Make node.Node create .server with p2p.NewServer() and use p2p.Server.AddProtocols() at .Start();
* Create lachesis/ package;