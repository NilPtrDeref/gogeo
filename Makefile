build:
	@cd internal/common && go generate
	@templ generate
	@go build -o gogeo .

setup:
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/tinylib/msgp@latest
