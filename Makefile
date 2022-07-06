gen:
	protoc --proto_path=proto proto/*.proto --go_out=server --go-grpc_out=server
	protoc --proto_path=proto proto/*.proto --go_out=es --go-grpc_out=es

clean:
	rm -rf server/pb/
	rm -rf client/pb/

server:
	go run server/main.go

install:
	go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

test:
	rm -rf tmp && mkdir tmp
	go test -cover -race serializer/*.go