SHELL = /bin/bash


.PHONY: help build release local all
.DEFAULT_GOAL := help
# From http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'

release: ## Build go app for alpine and publish to repo
	release --snapshot --skip-validate --rm-dist
	tar zxvf dist/*386.tar.gz
	cp vrctlviz rootfs/usr/bin/vrctlviz
	cp vrctlviz chart/rootfs/usr/bin/vrctlviz
	tar zxvf dist/*amd64.tar.gz
	git add . && git commit -a -m "auto update for binary release"
	git push
	rm /usr/bin/vrctlviz
	cp vrctlviz /usr/bin/vrctlviz
	chmod +x /usr/bin/vrctlviz

local: ## Build go app for alpine and publish to repo
	release --snapshot --skip-validate --rm-dist
	tar zxvf dist/*386.tar.gz
	cp vrctlviz rootfs/usr/bin/vrctlviz
	cp vrctlviz chart/rootfs/usr/bin/vrctlviz
	tar zxvf dist/*amd64.tar.gz
	docker build -t gcr.io/itc3-production/network-operator:master
	rm /usr/bin/vrctlviz
	cp vrctlviz /usr/bin/vrctlviz
	chmod +x /usr/bin/vrctlviz

artifacts: ## Create pipeline artifacts
	rm staging/kubernetes/*
	rm qa/kubernetes/*
	rm production/kubernetes/*
	helm install --name vrctlviz --namespace staging chart --debug --dry-run > staging/kubernetes/deploy.yaml
	helm install --name vrctlviz --namespace qa chart --debug --dry-run > qa/kubernetes/deploy.yaml
	helm install --name vrctlviz --namespace production chart --debug --dry-run > production/kubernetes/deploy.yaml
	sed '/apiVersion/,$!d' staging/kubernetes/deploy.yaml
	sed '/apiVersion/,$!d' qa/kubernetes/deploy.yaml
	sed '/apiVersion/,$!d' production/kubernetes/deploy.yaml


data: ## Update data file with contents from data folder
	go-bindata data/

all: ## Build go app, update s3 bucket, create alpine binary, update dockerfile rootfs and push repo for pipeline build
	make build
	make release

add:  ## add binary to path
	rm /usr/bin/vrctlviz
	cp ./vrctlviz /usr/bin/vrctlviz
	chmod +x /usr/bin/vrctlviz

docker: ## Install / reinstall docker
	chmod +x ./docker.install
	./docker.install
