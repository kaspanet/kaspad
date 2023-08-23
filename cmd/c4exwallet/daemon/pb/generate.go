//go:generate protoc --go_out=. --go-grpc_out=. --go_opt=paths=source_relative --go-grpc_opt=paths=source_relative c4exwalletd.proto

package pb
