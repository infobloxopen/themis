build:
	go build

.PHONY: clean
clean:
	go clean 

.PHONY: install
install:
	go install ./... 

.PHONY: lint
lint:
	go get -v github.com/golang/lint/golint
	if [ -n "$$(golint *.go)" ]; then \
		exit 1; \
	fi; \

.PHONY: vet
vet:
	go vet ./...
