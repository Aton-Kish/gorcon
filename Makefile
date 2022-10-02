include server/.env

.PHONY: help
help:
	@cat Makefile

.PHONY: start
start:
	docker run --rm -d --name minecraft -p $(RCON_PORT):$(RCON_PORT) --env-file server/.env -v ${PWD}/server/data:/data -v ${PWD}/server/mods.txt:/mods.txt itzg/minecraft-server

.PHONY: stop
stop:
	docker stop minecraft

.PHONY: unit
unit:
	go test ./...

.PHONY: e2e
e2e:
	go clean -testcache
	go test -tags e2e ./...
