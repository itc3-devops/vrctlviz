SHELL = /bin/bash


.PHONY: help build release local all
.DEFAULT_GOAL := help
# From http://marmelab.com/blog/2016/02/29/auto-documented-makefile.html
help:
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-30s\033[0m %s\n", $$1, $$2}'


bucket: ## Build go app for alpine and publish to repo
	env GOOS=linux GOARCH=386 go build -v
	gsutil cp collector gs://itc3-public/vrctlviz-alpine
	echo "wget https://storage.googleapis.com/itc3-public/vrctlviz-alpine && mv vrctlviz-alpine viz && chmod +x viz"
