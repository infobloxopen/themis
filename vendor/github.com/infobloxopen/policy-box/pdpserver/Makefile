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
	$(GOLINT) -set_exit_status ./...

.PHONY: vet
vet:
	go vet ./...
