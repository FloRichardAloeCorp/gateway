project_name:=$(shell grep "module " <go.mod | sed 's;.*/;;g')

full: upd-vendor format lint deploy

$(project_name):
	go build -a -mod=vendor -o ./$(@)

build_and_push_image:
	docker build . -t ghcr.io/beltsecurity/gateway:latest;
	docker push ghcr.io/beltsecurity/gateway:latest

build:
	$(MAKE) $(project_name)

format:
	go fmt ./...

upd-vendor:
	go mod tidy
	go mod vendor
 
test: upd-vendor 
	go clean -testcache
	GIN_MODE=release go test -timeout 1m -cover $$(go list ./... | grep -v test)

lint: 
	golangci-lint run --allow-parallel-runners -c ./.golangci-lint.yaml --fix ./...

godoc:
	command -v godoc >/dev/null 2>&1 || go get golang.org/x/tools/cmd/godoc
	echo -e "Go to http://localhost:6060/internal/$(shell head -n 1 go.mod | cut -d" " -f2)?m=all\nPress Ctrl + C to exit"
	godoc 

.PHONY: full $(project_name) deploy down build format upd-vendor test lint godoc