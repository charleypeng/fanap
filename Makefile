# Makefile for fanap

BINARY_NAME=fanap
BUILD_DIR=build
GO=go
GOFLAGS=-ldflags="-s -w"

.PHONY: all build clean install test run help

all: build

# 构建当前平台的二进制文件
build:
	@echo "构建 $(BINARY_NAME)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME) .

# 构建Linux x64版本
build-linux:
	@echo "构建 $(BINARY_NAME) for Linux x64..."
	@mkdir -p $(BUILD_DIR)
	GOOS=linux GOARCH=amd64 $(GO) build $(GOFLAGS) -o $(BUILD_DIR)/$(BINARY_NAME)-linux-amd64 .

# 构建所有平台版本
build-all: build build-linux

# 清理构建文件
clean:
	@echo "清理构建文件..."
	@rm -rf $(BUILD_DIR)

# 运行程序（需要在Linux环境下）
run:
	$(GO) run . -verbose

# 列出可用的传感器（需要在Linux环境下）
list:
	$(GO) run . -list

# 安装到系统
install:
	@echo "安装 $(BINARY_NAME)..."
	sudo cp $(BUILD_DIR)/$(BINARY_NAME) /usr/local/bin/
	sudo chmod +x /usr/local/bin/$(BINARY_NAME)

# 卸载
uninstall:
	@echo "卸载 $(BINARY_NAME)..."
	sudo rm -f /usr/local/bin/$(BINARY_NAME)

# 运行测试
test:
	$(GO) test -v ./...

# 格式化代码
fmt:
	$(GO) fmt ./...

# 代码检查
vet:
	$(GO) vet ./...

# 帮助信息
help:
	@echo "可用命令:"
	@echo "  make build       - 构建当前平台的二进制文件"
	@echo "  make build-linux - 构建Linux x64版本"
	@echo "  make build-all   - 构建所有平台版本"
	@echo "  make build-tools - 构建传感器列表工具"
	@echo "  make clean       - 清理构建文件"
	@echo "  make run         - 运行程序"
	@echo "  make list        - 列出可用的传感器"
	@echo "  make install     - 安装到系统"
	@echo "  make uninstall   - 从系统卸载"
	@echo "  make test        - 运行测试"
	@echo "  make fmt         - 格式化代码"
	@echo "  make vet         - 代码检查"
