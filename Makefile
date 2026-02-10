.PHONY: build

build:
	@cd internal/common && go generate
	@templ generate
	@mkdir -p build
	@go build -o 'build/gogeo' .

setup:
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/tinylib/msgp@latest

clean:
	@rm -rf build
