GO := go
ROOT_PACKAGE := github.com/marmotedu/marmotedu-sdk-go
ifeq ($(origin ROOT_DIR),undefined)    
ROOT_DIR := $(shell pwd)
endif    

all: test format lint boilerplate

## test: Test the package.
.PHONY: test
test:
	@echo "===========> Testing packages"
	@$(GO) test $(ROOT_PACKAGE)/...

## format: Format the package with `gofmt`
.PHONY: format
format:  
	@echo "===========> Formating codes"
	@find . -name "*.go" | xargs gofmt -s -w
	@find . -name "*.go" | xargs goimports -w -local $(ROOT_PACKAGE)

.PHONY: lint.verify                                                           
lint.verify:
ifeq (,$(shell which golangci-lint 2>/dev/null))
	@echo "===========> Installing golangci lint"
	@GO111MODULE=off $(GO) get -u github.com/golangci/golangci-lint/cmd/golangci-lint    
endif                       

## lint: Check syntax and styling of go sources.
.PHONY: lint    
lint: lint.verify
	@echo "===========> Run golangci to lint source codes"
	@golangci-lint run $(ROOT_DIR)/...  

.PHONY: license.verify    
license.verify:
	@echo "===========> Verifying the boilerplate headers for all files"
	@$(GO) run $(ROOT_DIR)/tools/addlicense/addlicense.go --check -f $(ROOT_DIR)/boilerplate.txt $(ROOT_DIR) --skip-dirs=third_party
    
.PHONY: license.add    
license.add:
	@$(GO) run $(ROOT_DIR)/tools/addlicense/addlicense.go -v -f $(ROOT_DIR)/boilerplate.txt $(ROOT_DIR) --skip-dirs=third_party

## boilerplate: Verify the boilerplate headers for all files.    
.PHONY: boilerplate    
boilerplate:
	@$(MAKE) license.verify                            
    
## license: Ensures source code files have copyright license headers.               
.PHONY: license    
license:
	@$(MAKE) license.add     

## help: Show this help info.
.PHONY: help
help: Makefile
	@echo -e "\nUsage: make <TARGETS> ...\n\nTargets:"
	@sed -n 's/^##//p' $< | column -t -s ':' |  sed -e 's/^/ /'
	@echo "$$USAGE_OPTIONS"
