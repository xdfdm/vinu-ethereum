.PHONY: all build clean start stop console TTY

all: build


build: docker/bin/geth docker/genesis.json
	#docker build -f ./Dockerfile.geth -t geth-with-lachesis ../../../../
	cd docker && docker build -t "geth-over-lachesis" -f Dockerfile.geth ./

docker/genesis.json: docker/default.pswd docker/bin/geth
	cd docker && for i in `seq 1 2`; do ./bin/geth --datadir ./gethdata --password default.pswd account new; done
	BALANCES="$$( cat docker/gethdata/keystore/UTC* | jq '.address' | sed 's/"\(.*\)"/\\"\1\\": { \\"balance\\": \\"1000000000000000000000000\\" },/g' | xargs | sed 's/.$$//' )" \
	&& sed "s/{}/{$$BALANCES}/" genesis.tmpl > docker/genesis.json
	cd docker && ./bin/geth --datadir ./gethdata init genesis.json

docker/bin/geth: docker/default.pswd
	go build -o docker/bin/geth ../../cmd/geth

docker/default.pswd: clean
	echo 123456 > docker/default.pswd

clean:
	find docker/ -type d -exec git clean -f -X {}/ \;



start:
	./docker/testnet-start.sh

stop:
	./docker/testnet-stop.sh



console1:
	docker exec -i geth1 /usr/bin/geth --datadir=/gethdata attach rpc:http://127.0.0.1:8545
console2:
	docker exec -i geth2 /usr/bin/geth --datadir=/gethdata attach rpc:http://127.0.0.1:8545

TTY1:
	docker exec -ti geth1 /usr/bin/geth --datadir=/gethdata attach rpc:http://127.0.0.1:8545
TTY2:
	docker exec -ti geth2 /usr/bin/geth --datadir=/gethdata attach rpc:http://127.0.0.1:8545
