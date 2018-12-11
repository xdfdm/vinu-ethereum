#!/bin/bash

set -ex

cd $(dirname $0)

N=2                                                                                                                                                                                                       

for i in $(seq 1 $N)
do
    docker stop geth$i
    docker rm geth$i
done                