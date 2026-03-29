.PHONY: install update test

install:
	go mod tidy

update:
	go get -u all
	go mod tidy
	gofmt -w -l .
	$(MAKE) test

test:
	go test ./gofiberswagger

EXAMPLES := auth-bearer basic custom-config enums file-upload manually-register-routes embedded-types
$(EXAMPLES):
	go run examples/$@/main.go
