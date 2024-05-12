.PHONY: clean build test run docker-* tag release

NAME=trading-gpt
VERSION=0.16.1

clean:
	rm -rf build/*

build: clean
	CGO_ENABLED=0 go build -o ./build/bbgo ./main.go

build-linux:
	CGO_ENABLED=0 GOOS=linux go build -o ./build/bbgo ./main.go

test:
	go test ./...

run: build
	./build/bbgo run --dotenv .env.local --config bbgo.yaml --lightweight false --no-sync false

docker-build: build-linux
	docker build --tag yubing744/${NAME}:${VERSION} .
	docker tag yubing744/${NAME}:${VERSION} yubing744/${NAME}:latest

docker-push:
	docker push yubing744/${NAME}:${VERSION}
	docker push yubing744/${NAME}:latest

docker-start:
	docker run --name ${NAME} --net host -d -v ${PWD}:/strategy yubing744/${NAME}:${VERSION} run --dotenv .env.local --config bbgo.yaml

docker-stop:
	docker rm -f ${NAME}

docker-logs:
	docker logs -f --tail 100 ${NAME}

tag:
	git tag -m "release v${VERSION}" v${VERSION}
	git push --tags

release: docker-build tag docker-push