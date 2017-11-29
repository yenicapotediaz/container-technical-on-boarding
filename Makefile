
PATH:="$(PATH):$(GOPATH)/bin/:`pwd`/go/bin/"
MODULES_PATH:="./github/issues/"
COMMAND_PATH:="./cmd/"

all: vet lint test build

build: prepare_workload

# TODO: use glide to populate vendored dependencies


setup: 
	@go version
	@echo GOPATH IS ${GOPATH}
	go get github.com/satori/go.uuid
	go get -u -fix github.com/google/go-github/github
	go get gopkg.in/yaml.v2
	go get golang.org/x/oauth2


prepare_workload: setup $(shell find $(MODULES_PATH) $(COMMAND_PATH) -name '*.go')
	go build -v -o $@  $(COMMAND_PATH)

test: setup vet lint
	go test -race -v $(MODULES_PATH)
	
coverage.html: $(shell find $(MODULES_PATH) $(COMMAND_PATH) -name '*.go')
	go test -covermode=count -coverprofile=coverage.prof $(MODULES_PATH)
	go tool cover -html=coverage.prof -o $@

test-cover: coverage.html

lint: setup
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u honnef.co/go/tools/cmd/gosimple
	gofmt -w -s ./github/issues/
	gofmt -w -s ./cmd/
	$(GOPATH)/bin/goimports -w ./github/issues/
	$(GOPATH)/bin/goimports -w ./cmd/
	$(GOPATH)/bin/golint ./github/issues/
	$(GOPATH)/bin/golint ./cmd/
	$(GOPATH)/bin/gosimple ./github/issues/
	$(GOPATH)/bin/gosimple ./cmd/
	
vet: 
	go vet -v $(MODULES_PATH)

clean:
	-rm -vf ./coverage.* ./prepare_workload


godoc.txt: $(shell find ./ -name '*.go')
	godoc $(MODULES_PATH) > $@

docs:  godoc.txt


.PHONY: vet lint test test-cover setup clean docs