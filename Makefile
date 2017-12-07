APP_NAME         = technical-on-boarding
GIT_VERSION      = $(shell git describe --tags --abbrev=0 2>/dev/null | sed 's/^v//')
APP_VERSION     ?= $(if $(GIT_VERSION),$(GIT_VERSION),0.0.0)
GIT_HASH         = $(shell git rev-parse --short HEAD)
APP_BUILD       ?= $(if $(GIT_HASH),$(GIT_HASH),000)
APP_PACKAGE      = github.com/samsung-cnct/container-technical-on-boarding
APP_PATH         = ./app
APP_PATH_PKGS    = $(APP_PATH)/models $(APP_PATH)/controllers $(APP_PATH)/jobs $(APP_PATH)/jobs/onboarding

IMAGE_REPO       = quay.io
IMAGE_REPO_ORG   = samsung_cnct
IMAGE_TAG       ?= local-dev
IMAGE_NAME      ?= $(IMAGE_REPO)/$(IMAGE_REPO_ORG)/$(APP_NAME):$(IMAGE_TAG)

LDFLAGS=-ldflags "-X main.Version=${APP_VERSION} -X main.Build=${APP_BUILD}"

DOCKER_RUN_OPTS  =--rm -it -p 9000:9000 --env-file ./.env
DOCKER_RUN_CMD  ?=

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

docker-build: Dockerfile
	docker build --pull --force-rm \
	   --build-arg VERSION=$(APP_VERSION) \
	   --build-arg BUILD=$(APP_BUILD) \
	   -t $(IMAGE_NAME) .
	touch docker-build

docker-test: docker-build
	docker run --rm --env-file ./template.env \
		 $(IMAGE_NAME) \
		 revel test $(APP_PACKAGE) dev

docker-run: docker-build
	docker run $(DOCKER_RUN_OPTS) $(IMAGE_NAME) $(DOCKER_RUN_CMD)

docker-run-dev: docker-build
	docker run $(DOCKER_RUN_OPTS) \
	   -v $(GOPATH):/go \
		 -e VERSION=$(APP_VERSION) \
		 -e BUILD=$(APP_BUILD) \
	   $(IMAGE_NAME) $(DOCKER_RUN_CMD)

docker-clean:
	rm docker-build
	docker rmi $(IMAGE_NAME)

.PHONY: vet lint test test-cover setup clean docs docker-test docker-run docker-run-dev docker-clean
