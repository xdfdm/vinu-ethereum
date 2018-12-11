#!/bin/bash

set -ex

cd $(dirname $0)

N=2                                                                                                                                                                                                       

for i in $(seq 1 $N)
do
    docker create --name=geth$i --net=lachesisnet --user $(id -u) geth-over-lachesis \
	--networkid 1313 \
	--rpc --rpcapi admin,eth,net,web3,personal \
	--lachesis=172.77.5.$i:1338 \
	--verbosity 6 \
	--nodiscover

    docker cp ./gethdata geth$i:/gethdata

    docker start geth$i
done                