# import configuration variables
include Makefile.env
export $(shell sed -E 's/\??=.*//' Makefile.env)

LDFLAGS=-ldflags "-X main.Version=${APP_VERSION} -X main.Build=${APP_BUILD}"
PATH:="$(PATH):$(GOPATH)/bin/:`pwd`/go/bin/"

all: vet lint test build

build: $(APP_NAME)

# TODO: use glide to populate vendored dependencies

setup:
	@go version
	@echo GOPATH IS ${GOPATH}
	@echo app.version is $(APP_VERSION)+$(APP_BUILD)
	go get github.com/satori/go.uuid
	go get -u -fix github.com/google/go-github/github
	go get gopkg.in/yaml.v2
	go get golang.org/x/oauth2
	go get github.com/revel/cmd/revel
	go get github.com/revel/revel
	go get github.com/revel/cron
	go get github.com/masterminds/semver

$(APP_NAME):
	go build -v $(LDFLAGS) $(APP_PATH_PKGS)

test: setup vet lint
	go test -race -v $(APP_PATH_PKGS)

coverage.html: $(shell find $(APP_PATH_PKGS) -name '*.go')
	go test -covermode=count -coverprofile=coverage.prof $(APP_PATH_PKGS)
	go tool cover -html=coverage.prof -o $@

test-cover: coverage.html

lint: setup
	go get -u github.com/golang/lint/golint
	go get -u golang.org/x/tools/cmd/goimports
	go get -u honnef.co/go/tools/cmd/gosimple
	gofmt -w -s $(APP_PATH_PKGS)
	$(GOPATH)/bin/goimports -w $(APP_PATH_PKGS)
	$(GOPATH)/bin/golint $(APP_PATH_PKGS)
	$(GOPATH)/bin/gosimple $(APP_PATH_PKGS)

vet:
	go vet -v -printf=false $(APP_PATH)

clean:
	-rm -vf ./coverage.* ./$(APP_NAME)
	-rm -rf ./test-results/

godoc.txt: $(shell find ./ -name '*.go')
	godoc $(APP_PATH) > $@

docs:  godoc.txt

.PHONY: vet lint test test-cover setup clean docs
