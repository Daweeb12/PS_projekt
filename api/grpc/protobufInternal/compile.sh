cd ~/fri/3letnik/1/ps/PS_projekt/
ls api/grpc/protobufInternal/
ls ~/fri/3letnik/1/ps/PS_projekt/api/grpc/

protoc -I=. --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative api/grpc/protobufInternal/protobufInt.proto
