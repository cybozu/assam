.DEFAULT_GOAL := help

.PHONY: help
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | \
		sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.PHONY: lint
lint: ## lint: Analyze code for potential errors.
	golangci-lint run

.PHONY: test
test: ## test: Test packages.
	mkdir -p test-results
	gotestsum --junitfile test-results/results.xml

.PHONY: package
package: ## package: build executable binary archives.
	goreleaser --snapshot --skip-publish --rm-dist
