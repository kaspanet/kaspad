serialization
=============

1. Download and place in your PATH: https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip
2. `go get github.com/golang/protobuf/protoc-gen-go`
3. `go get google.golang.org/grpc/cmd/protoc-gen-go-grpc`
4. In the protowire directory: `go generate .`