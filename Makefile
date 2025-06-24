BASE_OUTPUT="C:\Users\Lappy\OneDrive\Desktop\Cringe xD\Dim Future\Go\Picker-Pal\PKr-Base\PKr-Base.exe"

TEST_DEST="C:\Users\Lappy\OneDrive\Desktop\Cringe xD\Dim Future\Go\Picker-Pal\PKr-Test"
TEST_MOIT="C:\Users\Lappy\OneDrive\Desktop\Cringe xD\Dim Future\Go\Picker-Pal\PKr-Test\Moit"
TEST_PALAS="C:\Users\Lappy\OneDrive\Desktop\Cringe xD\Dim Future\Go\Picker-Pal\PKr-Test\Palas"

build2test:clean build copy done

done:
	@echo $(TEST) is built

build:
	@cls
	@echo Building the PKr-Base file...
	@go build -o PKr-Base.exe

copy:
	@echo Copying the executable to the destination...

	@copy $(BASE_OUTPUT) $(TEST_DEST)
	@copy $(BASE_OUTPUT) $(TEST_MOIT)
	@copy $(BASE_OUTPUT) $(TEST_PALAS)
	
	@del $(BASE_OUTPUT)

clean:
	@cls
	@echo Cleaning up...

	@del $(TEST_DEST)\PKr-base.exe || exit 0
	@del $(TEST_MOIT)\PKr-base.exe || exit 0
	@del $(TEST_PALAS)\PKr-base.exe || exit 0

grpc-out:
	protoc ./proto/*.proto --go_out=. --go-grpc_out=.

get-new-kcp:
	go get github.com/ButterHost69/kcp-go@latest

generate_icon:
	go install github.com/akavel/rsrc@latest
	rsrc -ico .\PKrBase.ico -o PKrBase.syso

