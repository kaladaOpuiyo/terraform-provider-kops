.PHONY: help

help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

.DEFAULT_GOAL := help

## Dependency issues with no existent pkgs cant use glide install yet
## github.com/miekg/coredns/middleware/etcd/msg - seems to not exist?
## github.com/digitalocean/godo/context - seems to not exist?
build: ## Build Provider
	dep ensure -v && \
	cp  ./models/bindata.go ./vendor/k8s.io/kops/upup/models && \
	go build -o ~/.terraform.d/plugins/darwin_amd64/terraform-provider-kops
