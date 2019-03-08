#!/bin/bash

set -ex

cd $(dirname $0)

N=5                                                                                                                                                                                                       

for i in $(seq 1 $N)
do
    docker create --cpus="0.3" --name=geth$i --net=lachesisnet --user $(id -u) geth-over-lachesis \
	--networkid 1313 \
	--rpc --rpcapi "admin,eth,net,web3,personal,mine,debug,txpool" \
	--lachesis=172.31.29.148:9000 \
	--verbosity 6 \
	--nodiscover \
	--ws --wsport 8546 --wsorigins="*" \
	--syncmode "full" --gcmode "archive" \
	--mine --miner.gasprice=0 --miner.threads=1 --ethash.dagdir=/gethdata

# --fakepow

    docker cp ./gethdata geth$i:/gethdata

    docker start geth$i
done                