
# Create version tag from git tag
VERSION=$(shell git describe | sed 's/^v//')
REPO=cybermaggedon/evs-riskgraph
DOCKER=docker
GO=GOPATH=$$(pwd)/go go

all: evs-riskgraph build

SOURCE=evs-riskgraph.go config.go model.go gaffer.go domain.go

evs-riskgraph: ${SOURCE} go.mod go.sum
	${GO} build -o $@ ${SOURCE}

build: evs-riskgraph
	${DOCKER} build -t ${REPO}:${VERSION} -f Dockerfile .
	${DOCKER} tag ${REPO}:${VERSION} ${REPO}:latest

push:
	${DOCKER} push ${REPO}:${VERSION}
	${DOCKER} push ${REPO}:latest

