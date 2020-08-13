.PHONY: proto

proto:
	protoc --proto_path=./dnsseed/pb --go_out=./dnsseed/pb dnsseed/pb/peer_service.proto
	protoc --proto_path=./dnsseed/pb --go-grpc_out=./dnsseed/pb dnsseed/pb/peer_service.proto
	mv dnsseed/pb/github.com/kaspanet/kaspad/pb/*.go dnsseed/pb/
	rm -rf dnsseed/pb/github.com