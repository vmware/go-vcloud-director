TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go')
maindir=$(PWD)

default: fmtcheck vet static security build

# test runs the test suite and vets the code
test: testunit tagverify
	@echo "==> Running Functional Tests"
	cd govcd && go test -tags "functional" -timeout=650m -check.vv

# tagverify checks that each tag can run independently
tagverify: fmtcheck
	@echo "==> Running Tags Tests"
	@./scripts/test-tags.sh

# testunit runs the unit tests
testunit: fmtcheck
	@echo "==> Running Unit Tests"
	cd $(maindir)/govcd && go test -tags unit -v
	cd $(maindir)/util && go test -v

# testrace runs the race checker
testrace:
	@go list $(TEST) | xargs -n1 go test -race $(TESTARGS)

# This will include tests guarded by build tag concurrent with race detector
testconcurrent:
	cd govcd && go test -race -tags "api concurrent" -timeout 15m -check.vv -check.f "Test.*Concurrent"

# tests only catalog related features
testcatalog:
	cd govcd && go test -tags "catalog" -timeout 15m -check.vv

# tests only vapp and vm features
testvapp:
	cd govcd && go test -tags "vapp vm" -timeout 25m -check.vv

# tests only edge gateway features
testgateway:
	cd govcd && go test -tags "gateway" -timeout 15m -check.vv

# tests only networking features
testnetwork:
	cd govcd && go test -tags "network" -timeout 15m -check.vv

# tests only load balancer features
testlb:
	cd govcd && go test -tags "lb" -timeout 15m -check.vv

# tests only NSXV related features
testnsxv:
	cd govcd && go test -tags "nsxv" -timeout 15m -check.vv

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@echo "==> Running Go Vet"
	@go vet -tags ALL ./... ; if [ $$? -ne 0 ] ; then echo "vet error!" ; exit 1 ; fi

# static runs the source code static analysis tool `staticcheck`
static: fmtcheck
	@./scripts/staticcheck.sh

# security runs the source code security analysis tool `gosec`
security: fmtcheck
	@./scripts/gosec.sh

get-deps:
	@echo "==> Fetching dependencies"
	@go get -v $(TEST)
	@go get -u github.com/golang/lint/golint

fmt:
	gofmt -w $(GOFMT_FILES)

fmtcheck:
	@sh -c "'$(CURDIR)/scripts/gofmtcheck.sh'"

copyright:
	@echo "==> Checking copyright headers in source files"
	@sh -c "'$(CURDIR)/scripts/copyright_check.sh'"

build:
	@echo "==> Building govcd library"
	cd govcd && go build . && go test -tags ALL -c

