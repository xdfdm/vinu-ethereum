#!/bin/bash

set -x
cd $(dirname $0)

N=${1:-1}
PSWD=`cat docker/default.pswd`

make console$N << JAVASCRIPT

  personal.unlockAccount(eth.accounts[$N-1], "${PSWD}");
  eth.sendTransaction( {from:eth.accounts[$N-1], to:eth.accounts[2-$N], value: web3.toWei(1.0, "ether")} );


JAVASCRIPT
