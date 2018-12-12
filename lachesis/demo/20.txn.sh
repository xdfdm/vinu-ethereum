#!/bin/bash

set -x
cd $(dirname $0)

N=${1:-1}
PSWD=`cat docker/default.pswd`

make console$N << JAVASCRIPT

  personal.unlockAccount(eth.accounts[$N], "${PSWD}");
  eth.sendTransaction( {from:eth.accounts[$N], to:eth.accounts[3-$N], value: web3.toWei(1.0, "ether"), gasPrice: web3.toWei(40,'gwei')} );


JAVASCRIPT
