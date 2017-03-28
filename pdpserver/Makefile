ifndef GOPATH
$(error GOPATH isn\'t set)
else
$(info Using GOPATH=$(GOPATH))
endif

GOLINT ?= $(GOPATH)/bin/golint

$(GOLINT):
	go get -v github.com/golang/lint/golint

build:
	go build

.PHONY: clean
clean:
	go clean 

.PHONY: install
install:
	go install ./... 

.PHONY: lint
lint: $(GOLINT)
	if [ -n "$$(golint ./...)" ]; then \
		exit 1; \
	fi; \

.PHONY: vet
vet:
	go vet ./...
