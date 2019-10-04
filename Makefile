all:
	protoc -I/usr/local/include -I.  --go_out=plugins=grpc:.  protoexpl/data.proto
	protoc -I/usr/local/include -I.  --grpc-gateway_out=logtostderr=true:. protoexpl/data.proto
	protoc -I/usr/local/include -I.  --swagger_out=logtostderr=true:.  protoexpl/data.proto
	protoc-go-inject-tag -input=./protoexpl/data.pb.go
	go generate .