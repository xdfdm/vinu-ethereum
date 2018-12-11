#!/bin/bash

set -e
cd $(dirname $0)

N=${1:-1}

make console$N << JAVASCRIPT | egrep "^#"


function checkAllBalances() {
  var i = 0;
  eth.accounts.forEach( function(e){
    console.log("#  eth.accounts["+i+"]: " +  e + " \tbalance: " + web3.fromWei(eth.getBalance(e), "ether") + " ether");
    i++;
  } );

  eth.getBlock("pending", true).transactions.forEach( function(e){
    console.log("#  pending txn: ", e );
  } );

};
checkAllBalances();


JAVASCRIPT
