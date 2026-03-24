.PHONY: all
all: init-env
	go run ./...

.PHONY: init-env
init-env:
	op inject -i .envrc.example -o .envrc


.PHONY: clean
clean:
	rm rttui

.PHONY: test
test:
	go test ./...
