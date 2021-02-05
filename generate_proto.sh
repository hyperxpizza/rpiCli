export PATH=$PATH:$HOME/go/bin
export PATH=$PATH:/usr/local/go/bin
protoc grpc/protos.proto --go_out=plugins=grpc:. --go_opt=paths=source_relative