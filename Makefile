.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help


build: ## Build Provider ¯\_(ツ)_/¯
	dep ensure -v && \
	cp  ./models/bindata.go ./vendor/k8s.io/kops/upup/models && \
	go build -o ~/.terraform.d/plugins/darwin_amd64/terraform-provider-kops
