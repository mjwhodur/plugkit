default: pregit

test: tm build 0-plugin-test 1-plugin-test 2-plugin-test 3-plugin-test

tm:
	@echo "Beginning standard testing procedure"

clean:
	@rm ./host
	@rm ./plugin
	@echo "Repo cleared"

build:
	@go build .
	@echo "Library build finished. No error reported."

0-plugin-test:
	@echo "==== 0-plugin-test procedure ===="
	@go build -o host ./examples/0-smartplug-test-basic/client
	@go build -o plugin ./examples/0-smartplug-test-basic/plug
	./host
	@echo
	@echo "No error reported."

1-plugin-test:
	@echo "==== 1-plugin-test procedure ===="
	@go build -o host ./examples/1-rawclient-smartplug-test-basic/client
	@go build -o plugin ./examples/1-rawclient-smartplug-test-basic/plug
	./host
	@echo
	@echo "No error reported."

2-plugin-test:
	@echo "==== 2-plugin-test procedure ===="
	@go build -o host ./examples/2-rawclient-rawplug-test-basic/client
	@go build -o plugin ./examples/2-rawclient-rawplug-test-basic/plug
	./host
	@echo
	@echo "No error reported."

3-plugin-test:
	@echo "==== 2-plugin-test procedure ===="
	@go build -o host ./examples/3-rawstream-basic/client
	@go build -o plugin ./examples/3-rawstream-basic/plug
	./host
	@echo
	@echo "No error reported."

BIN_DIR := bin
GOLANGCI_LINT := $(BIN_DIR)/golangci-lint

$(GOLANGCI_LINT):
	@echo "🔧 Golangci-lint not found - downloading..."
	@mkdir -p $(BIN_DIR)
	@curl curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest
	@echo "✅ Installed golangci-lint."

.PHONY: lint
lint: $(GOLANGCI_LINT)
	@$(GOLANGCI_LINT) run

update-golangci:
	@curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest

format:
	@$(GOLANGCI_LINT) run --fix

release-notes:
	git-cliff > CHANGELOG.md

pregit: format test clean release-notes

cargo-brew:
	brew install rust

cliff-install:
	cargo install git-cliff