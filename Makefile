.PHONY: clean sync backtest build test run docker-build docker-push docker-run start stop logs deploy

NAME=trading-bot
VERSION=0.2.7-beta
TARGET_KEY=-i ~/.ssh/earn-robot_key.pem
DEPLOY_TARGET=azureuser@20.59.104.24
DEPLOY_PATH=/home/azureuser/apps/quant/bbgo-strategys/${NAME}

clean:
	rm -rf build/*

sync:
	go run ./cmd/bbgo.go sync

backtest:
	go run ./cmd/bbgo.go backtest -v --sync --config ./bbgo.yaml

build: clean
	CGO_ENABLED=0 go build -o ./build/bbgo ./cmd/bbgo.go

test:
	go test ./...

run: build
	./build/bbgo run

docker-build: build
	docker build --tag yubing744/${NAME}:latest .
	docker tag yubing744/${NAME}:latest yubing744/${NAME}:${VERSION}

docker-push: docker-build
	docker push yubing744/${NAME}:${VERSION}

docker-run:
	docker run --net host -v ${PWD}:/strategy yubing744/${NAME}:${VERSION} run

start:
	docker run --name ${NAME} --net host -d -v ${PWD}:/strategy yubing744/${NAME}:${VERSION} run

stop:
	docker rm -f ${NAME}

logs:
	docker logs -f --tail 100 ${NAME}

deploy: docker-push
	ssh ${TARGET_KEY} ${DEPLOY_TARGET} "mkdir -p ${DEPLOY_PATH}"
	scp ${TARGET_KEY} .env.local bbgo.yaml Makefile ${DEPLOY_TARGET}:${DEPLOY_PATH}
	ssh ${TARGET_KEY} ${DEPLOY_TARGET} "cd ${DEPLOY_PATH} && make stop && make start && make logs"

remote-logs:
	ssh ${TARGET_KEY} ${DEPLOY_TARGET} "cd ${DEPLOY_PATH} && make logs"

remote-shell:
	ssh ${TARGET_KEY} ${DEPLOY_TARGET}
