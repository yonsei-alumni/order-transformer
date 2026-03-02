APP_NAME := order-transformer
DIST_DIR := dist

.PHONY: all linux windows darwin clean

all: linux windows

linux:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=linux GOARCH=amd64 go build -o $(DIST_DIR)/$(APP_NAME)-linux-amd64 .

windows:
	@mkdir -p $(DIST_DIR)
	CGO_ENABLED=1 GOOS=windows GOARCH=amd64 CC=x86_64-w64-mingw32-gcc go build -o $(DIST_DIR)/$(APP_NAME)-windows-amd64.exe .

darwin:
	@echo "macOS 빌드는 Mac에서 실행하세요: go build -o $(APP_NAME) ."

clean:
	rm -rf $(DIST_DIR)
