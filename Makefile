.PHONY: clean build test run docker-*

NAME=trading-gpt
VERSION=0.10.1-beta1

clean:
	rm -rf build/*

build: clean
	CGO_ENABLED=0 go build -o ./build/bbgo ./cmd/bbgo.go

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -o ./build/bbgo ./cmd/bbgo.go

test:
	go test ./...

run: build
	./build/bbgo run --dotenv .env.local --config bbgo.yaml

docker-build: build-linux
	docker build --tag yubing744/${NAME}:latest .
	docker tag yubing744/${NAME}:latest yubing744/${NAME}:${VERSION}

docker-push: docker-build
	docker push yubing744/${NAME}:${VERSION}
	docker push yubing744/${NAME}:latest

docker-start:
	docker run --name ${NAME} --net host -d -v ${PWD}:/strategy yubing744/${NAME}:${VERSION} run --dotenv .env.local --config bbgo.yaml

docker-stop:
	docker rm -f ${NAME}

docker-logs:
	docker logs -f --tail 100 ${NAME}
