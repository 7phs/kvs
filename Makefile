IMAGE = github.com/7phs/kvs
VERSION = latest

.PHONY: build
build:
	go build -o ./bin/kvs ./cmd/kvs

.PHONY: run
run:
	go run ./cmd/kvs

.PHONY: test
test:
	LOG_LEVEL=error go test --race ./...

.PHONY: test-integrations
test-integrations:
	docker-compose -f docker-compose.yml \
                   -f docker-compose.test.integrations.yml \
	               up --build --abort-on-container-exit
.PHONY: image
image:
	docker build -t $(IMAGE):$(VERSION)  .

.PHONY: image-run
image-run:
	docker run --rm -it -p 9889:9889 $(IMAGE):$(VERSION)

.PHONY: push
push:
	docker push $(IMAGE):$(VERSION)
