APP_DIR=${CURDIR}
APP_BIN:=bin
APP_SRC:=src

PACKAGE_NAME:=app/simulator
TOOL_NAME:=simulator

GOFMT:=gofmt
GOPATH:=${APP_DIR}
GO:=env GOPATH="${GOPATH}" go

.PHONY: clean build

all: clean fmt build

clean:
	@echo "Cleaning build directory..."
	@rm -rf $(APP_DIR)/$(APP_BIN)

fmt:
	@echo "Formating app..."
	@$(GOFMT) -l -w $(APP_DIR)/$(APP_SRC)/$(PACKAGE_NAME)

build: fmt
	@echo "Building app..."
	@$(GO) build -o $(APP_DIR)/$(APP_BIN)/$(TOOL_NAME).exe $(PACKAGE_NAME)

test: build
	@echo "Testing app..."
	@$(GO) test $(PACKAGE_NAME)...