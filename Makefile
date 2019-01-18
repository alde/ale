.PHONY: all test test-server test-docker docker docker-clean publish-docker

REPO=github.com/alde/ale
VERSION?=$(shell git describe HEAD --always | sed s/^v//)
DATE?=$(shell date -u '+%Y-%m-%d_%H:%M:%S')
DOCKERNAME?=alde/ale
DOCKERTAG?=${DOCKERNAME}:${VERSION}
LDFLAGS=-X ${REPO}/version.Version=${VERSION} -X ${REPO}/version.BuildDate=${DATE}
SRC=$(shell find . -name '*.go')
TESTFLAGS="-v"

DOCKER_GO_SRC_PATH=/go/src/github.com/alde/ale
DOCKER_GOLANG_RUN_CMD=docker run --rm -v "$(PWD)":$(DOCKER_GO_SRC_PATH) -w $(DOCKER_GO_SRC_PATH) golang:1.11 bash -c

PACKAGES=$(shell go list ./... | grep -v /vendor/)

all: test

test: ale
	go test $(shell glide novendor)

coverage:
	echo "mode: count" > coverage-all.out
	$(foreach pkg,$(PACKAGES),\
		go test -coverprofile=coverage.out -covermode=count $(pkg);\
		tail -n +2 coverage.out >> coverage-all.out;)
	go tool cover -html=coverage-all.out

# Run tests cleanly in a docker container.
test-docker:
	$(DOCKER_GOLANG_RUN_CMD) "make test"


vet:
	go vet ${PACKAGES}

lint:
	go list ./... | grep -v /vendor/ | grep -v assets | xargs -L1 golint -set_exit_status

ale: ${SRC}
	go build -ldflags "${LDFLAGS}" -o $@ github.com/alde/ale/cmd/ale

docker/ale: ${SRC}
	CGO_ENABLED=0 GOOS=linux go build -ldflags "${LDFLAGS}" -a -installsuffix cgo -o $@ github.com/alde/ale/cmd/ale

docker: docker/ale docker/Dockerfile
	docker build -t ${DOCKERTAG} docker

docker-clean: docker/Dockerfile
	# Create the docker/ale binary in the Docker container using the
	# golang docker image. This ensures a completely clean build.
	$(DOCKER_GOLANG_RUN_CMD) "make docker/ale"
	docker build -t ${DOCKERTAG} docker

publish-docker:
#ifeq ($(strip $(shell docker images --format="{{.Repository}}:{{.Tag}}" $(DOCKERTAG))),)
#	$(warning Docker tag does not exist:)
#	$(warning ${DOCKERTAG})
#	$(warning )
#	$(error Cannot publish the docker image. Please run `make docker` or `make docker-clean` first.)
#endif
	docker push ${DOCKERTAG}
	git describe HEAD --exact 2>/dev/null && \
		docker tag ${DOCKERTAG} ${DOCKERNAME}:latest && \
		docker push ${DOCKERNAME}:latest || true
