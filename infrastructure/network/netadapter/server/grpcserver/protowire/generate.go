//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative p2p.proto rpc.proto messages.proto
//go:generate protoc --doc_out=. --doc_opt=markdown,rpc.md rpc.proto

package protowire
