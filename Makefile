PROJECT_NAME := field
SERVICE_NAME := field-service
SERVICE_REPO := localhost:32000/olympsis-field:registry
PKG := "olympsis-services/$(PROJECT_NAME)"
PKG_LIST := $( go list ${PKG}/... | grep -v /vendor/)
GO_FILES := $( find . -name '*.go' | grep -v /vendor/ | grep -v _test.go)

.PHONY: all dep build clean test coverage coverhtml lint

all: build

lint: ## Lint the files
	golint -set_exit_status ${PKG_LIST}

test: ## Run unittests
	go test -short ${PKG_LIST}

race: dep ## Run data race detector
	go test -race -short ${PKG_LIST}

msan: dep ## Run memory sanitizer
	go test -msan -short ${PKG_LIST}

dep: ## Get the dependencies
	go get -v -d ./...

build: dep ## Build the binary file
	go build -v $(PKG) 

docker:
	docker build . -t $(SERVICE_REPO)
	docker rmi $$(docker images -f "dangling=true" -q) --force
	docker tag $$(docker images --filter 'reference=localhost:32000/olympsis-field' --format "{{.ID}}") $(SERVICE_REPO)
	docker push $(SERVICE_REPO)

new-docker:
	docker build . -t $(SERVICE_REPO)
	docker tag $$(docker images --filter 'reference=localhost:32000/olympsis-field' --format "{{.ID}}") $(SERVICE_REPO)
	docker push $(SERVICE_REPO)

run:
	docker run -p 7002:7002 localhost:32000/olympsis-field:registry 

clean: ## Remove previous build
	rm -f $(PROJECT_NAME)
