.PHONY: build


build:
	docker build -f ./Dockerfile.geth -t geth-lachesis ../../../../

start:
	docker run --rm --name=geth1 --net=lachesisnet -it geth-lachesis \
	-lachesis=172.77.5.1:1338 \
	|| echo "Start github.com/Fantom-foundation/go-lachesis/demo lachesistestnet first!"
