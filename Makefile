.PHONY: init-env
init-env:
	op inject -i .envrc.example -o .envrc

.PHONY: all
all:
	go run ./...

.PHONY: clean
clean:
	rm rttui

.PHONY: test
test:
	go test ./...
