protoc -I. --go_out=plugins=grpc:./pb *.proto
#protoc --grpc-gateway_out=logtostderr=true:./pb/ ./kube_opt.proto.proto
#protoc  --swagger_out=logtostderr=true:../entity/entity_pb ./widget.proto