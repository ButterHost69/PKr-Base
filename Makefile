protofiles:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.


protoc_files:
	protoc --go_out=. ./proto/*.proto

build_base:
	DEL /S PKr-base.exe && go build
# Prefered One
# .PHONY protofiles