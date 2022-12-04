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

.PHONY: unittest
unittest:
	go test ./...

.PHONY: e2etest
e2etest:
	go clean -testcache
	go test -tags e2e ./...

.PHONY: start
start:
	docker run --rm -d --name minecraft -p $(RCON_PORT):$(RCON_PORT) --env-file server/.env -v ${PWD}/server/data:/data -v ${PWD}/server/mods.txt:/mods.txt itzg/minecraft-server
	@while [ $$(docker logs minecraft | grep "Thread RCON Listener started" | wc -l) -eq 0 ]; do \
		echo -n "\rRCON is starting up"; \
		sleep 1; \
	done; \
	echo "\rRCON is started    "

.PHONY: stop
stop:
	docker stop minecraft

.PHONY: tcpdump
tcpdump:
	sudo tcpdump -X port 25575 -t

.PHONY: clean
clean:
	go mod tidy
	go clean --modcache
	@rm -rf $(shell pwd)/server/data
	@make uninstall
