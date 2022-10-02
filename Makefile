include server/.env

.PHONY: help
help:
	@cat Makefile

.PHONY: install
install:
	go install golang.org/x/tools/cmd/godoc@latest

.PHONY: uninstall
uninstall:
	rm $(shell go env GOPATH)/bin/godoc

.PHONY: doc
doc:
	godoc -http ":6060"

.PHONY: unit
unit:
	go test ./...

.PHONY: e2e
e2e:
	go clean -testcache
	go test -tags e2e ./...

.PHONY: start
start:
	docker run --rm -d --name minecraft -p $(RCON_PORT):$(RCON_PORT) --env-file server/.env -v ${PWD}/server/data:/data -v ${PWD}/server/mods.txt:/mods.txt itzg/minecraft-server

.PHONY: stop
stop:
	docker stop minecraft

.PHONY: clean
clean:
	go mod tidy
	go clean --modcache
	@rm -rf $(shell pwd)/server/data
	@make uninstall
