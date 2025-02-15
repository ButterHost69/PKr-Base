protofiles:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.


protoc_files:
	protoc --go_out=. ./proto/*.proto

get_imports:
	go get github.com/ccding/go-stun

build_base:
	DEL /S PKr-base.exe && go build
# Prefered One
# .PHONY protofiles