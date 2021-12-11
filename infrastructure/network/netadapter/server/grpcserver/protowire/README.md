protowire
=========

1. Download and place in your PATH:
   https://github.com/protocolbuffers/protobuf/releases/download/v3.12.3/protoc-3.12.3-linux-x86_64.zip
2. `go install github.com/golang/protobuf/protoc-gen-go`
3. `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc`
4. In the protowire directory: `go generate .`

Documentation
-------------r

To generate `rpc.md`:
1. `go install -u github.com/kaspanet/protoc-gen-doc/cmd/protoc-gen-doc`
2. In the protowire directory: `protoc --doc_out=. --doc_opt=markdown,rpc.md rpc.proto`
