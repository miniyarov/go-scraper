DIST_DIR := ./dist
REPORT_DIR := $(DIST_DIR)/report

.PHONY: test
test:
	@echo "Running unit tests"
	mkdir -p $(REPORT_DIR)
	go test -v -covermode=count -coverprofile=$(REPORT_DIR)/coverage.out -failfast ./...

.PHONY: coverage
coverage: test
	go tool cover -html=$(REPORT_DIR)/coverage.out -o $(REPORT_DIR)/coverage.html

.PHONY: integration
integration:
	go test -v -tags=integration

.PHONY: all
all: test coverage integration