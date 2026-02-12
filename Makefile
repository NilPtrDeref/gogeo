.PHONY: build

build:
	@go generate ./...
	@mkdir -p build
	@go build -o 'build/gogeo' .

serve: build
	@build/gogeo serve

setup:
	@cd frontend && npm i
	@go install github.com/air-verse/air@latest
	@go install github.com/tinylib/msgp@latest

clean:
	@rm -rf build
