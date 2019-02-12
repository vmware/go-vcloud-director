TEST?=./...
GOFMT_FILES?=$$(find . -name '*.go')

default: fmtcheck vet build

# test runs the test suite and vets the code
test: fmtcheck
	@echo "==> Running Tests"
	cd govcd && go test -timeout=45m -check.vv .

# testrace runs the race checker
testrace:
	@go list $(TEST) | xargs -n1 go test -race $(TESTARGS)

# vet runs the Go source code static analysis tool `vet` to find
# any common errors.
vet:
	@echo "==> Running Go Vet"
	@cd govcd && go vet ; if [ $$? -ne 0 ] ; then echo "vet error!" ; exit 1 ; fi && cd -

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
	cd govcd && go build . && go test -c .

