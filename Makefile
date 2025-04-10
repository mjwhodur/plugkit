default: test

test: tm build 0-plugin-test clean

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
	@go build -o host ./examples/0-plugin-test/client
	@go build -o plugin ./examples/0-plugin-test/plug
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
	@curl curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s latest

format:
	@$(GOLANGCI_LINT) run --fix

pregit: format test