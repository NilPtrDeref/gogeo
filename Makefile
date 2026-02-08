build:
	# Go protobufs
	@rm -rf internal/common/pb
	@protoc --go_out=internal/common --go_opt=paths=source_relative proto/message.proto
	@mv -f internal/common/proto internal/common/pb
	# JS protobufs
	@rm -rf cmd/serve/static/proto
	@mkdir -p cmd/serve/static/proto
	@protoc --js_out=import_style=closure,binary:cmd/serve/static/proto proto/message.proto
	# Html templates
	@templ generate
	# Binary
	@go build -o gogeo .

setup:
	go install github.com/air-verse/air@latest
	go install github.com/a-h/templ/cmd/templ@latest
	go install github.com/a-h/templ/cmd/templ@latest
	npm install --global @protocolbuffers/protoc-gen-js
