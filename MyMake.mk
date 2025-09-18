ROOT_DIR=E:\Projects\Picker-Pal

BASE_OUTPUT=$(ROOT_DIR)\PKr-Base\PKr-Base.exe
TEST_DEST=$(ROOT_DIR)\PKr-Test
TEST_MOIT=$(TEST_DEST)\Moit
TEST_PALAS=$(TEST_DEST)\Palas

build2test:clean build copy

build:
	@cls
	@echo Building the PKr-Base File ...
	@go build -o PKr-Base.exe

copy:
	@echo Copying the Executable to Test Destination ...

	@copy "$(BASE_OUTPUT)" "$(TEST_MOIT)"
	@copy "$(BASE_OUTPUT)" "$(TEST_PALAS)"
	
	@del "$(BASE_OUTPUT)"

clean:
	@cls
	@echo Cleaning Up ...

	@del "$(TEST_MOIT)\PKr-Base.exe" || exit 0
	@del "$(TEST_PALAS)\PKr-Base.exe" || exit 0
