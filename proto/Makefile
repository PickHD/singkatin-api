.PHONY: generate

generate:
	protoc --go_out=. --go_opt=paths=source_relative \
	       --go-grpc_out=. --go-grpc_opt=paths=source_relative \
	       api/v1/proto/shortener/shortener.proto \
	       api/v1/proto/upload/upload.proto
