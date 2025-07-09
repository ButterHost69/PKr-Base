grpc-out:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.

get-new-kcp:
	go get github.com/ButterHost69/kcp-go@latest

generate_icon:
	go install github.com/akavel/rsrc@latest
	rsrc -ico .\PKrBase.ico -o PKrBase.syso

generate_exe_with_no_terminal:
	go build -ldflags -H=windowsgui -o NoTerminal.exe
