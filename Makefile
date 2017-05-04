THEMIS=github.com/infobloxopen/themis
TESTABLE=pdp pep contrib/coredns/policy
BUILDABLE=$(TESTABLE) papcli pdp-control pdpctrl-client pdpserver pdp-service pepcli

BUILD_VERBOSE=-v
TEST_VERBOSE=-v

all: check

.PHONY: build
build:
	for i in $(BUILDABLE); do \
		(echo $$i; cd $$i; go build $(BUILD_VERBOSE)) \
	done

fmt:
	## run go fmt
	@test -z "$$(gofmt -s -l $(BUILDABLE) | tee /dev/stderr)" || \
		(echo "please format Go code with 'gofmt -s -w'" && false)

.PHONY: check
check: fmt build

.PHONY: test
test: check
	for i in $(TESTABLE); do \
	 go test $(TEST_VERBOSE) $(THEMIS)/$$i; \
	done

.PHONY: coverage
coverage: check
	rm -f coverage.txt
	for i in $(TESTABLE); do \
	 go test $(TEST_VERBOSE) -race -coverprofile=cover.out -covermode=atomic $(THEMIS)/$$i; \
	 cat cover.out >> coverage.txt ; \
	done
