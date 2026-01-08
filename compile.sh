cd ~/fri/3letnik/1/ps/PS_projekt

# Generate protobufRazpravljalnica
protoc -I. \
  --go_out=. \
  --go_opt=module=PS_projekt \
  --go-grpc_out=. \
  --go-grpc_opt=module=PS_projekt \
  api/grpc/protobufRazpravljalnica/protobuf.proto


