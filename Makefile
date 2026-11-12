BINARY_NAME=abby
ifeq ($(OS),Windows_NT)
	BINARY_EXT=.exe
endif

.PHONY: build test clean install help

build: ## Compile le binaire CLI dans ./bin/abby(.exe)
	go build -o bin/$(BINARY_NAME)$(BINARY_EXT) ./cmd/abby

test: ## Exécute les tests unitaires Go
	go test -v -race ./...

install: ## Installe le binaire dans $GOPATH/bin
	go install ./cmd/abby

clean: ## Nettoie les binaires compilés
	rm -rf bin/
