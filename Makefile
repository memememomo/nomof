GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
GOFMT=$(GOCMD) fmt
DOCKER_GO=docker-compose run go
DEPS=dep ensure

BINARY_NAME=nomof

all: test build

build:
	$(GOBUILD) -o $(BINARY_NAME) -v

test:
	$(GOTEST) -v ./...

test-dynamo:
	$(DOCKER_GO) make test

clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)

run:
	$(GOBUILD) -o $(BINARY_NAME) -v ./...

deps:
	$(DOCKER_GO) $(DEPS)

fmt:
	$(DOCKER_GO) $(GOFMT) ./...
