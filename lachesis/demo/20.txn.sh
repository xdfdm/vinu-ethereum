#!/bin/bash

set -x
cd $(dirname $0)

PSWD=`cat docker/default.pswd`

make console << JAVASCRIPT


  personal.unlockAccount(eth.accounts[0], "${PSWD}");
  eth.sendTransaction( {from:eth.accounts[0], to:eth.accounts[1], value: web3.toWei(1.0, "ether")} );


JAVASCRIPT
