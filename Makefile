#
# Makefile for Mattermost Giphy Plugin
#

# REMEMBER to set the OS and ARCHITECTURE of the
# Mattermost server for the Go cross-compilation :
TARGET_OS=linux
TARGET_ARCH=amd64

DEP_VERSION=0.4.1

SRC=plugin.go gifProvider.go
EXEC=plugin
CONF=plugin.yaml
PACKAGE=plugin.tar.gz
TEST=plugin_test.go gifProvider.go

all: test dist

$(EXEC): $(SRC)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -o $(EXEC) $(SRC)

build: $(EXEC)

rebuild: clean build

dist: $(EXEC) $(CONF)
	rm -rf ./dist
	mkdir ./dist
	@echo "BEWARE, if this command is executed on Windows, the executable bit of the plugin executable will NOT be set correctly..."
	chmod a+x $(EXEC) && tar -czvf dist/$(PACKAGE) $(EXEC) $(CONF)

test: $(SRC) $(TEST)
	go test .

vendor: Gopkg.lock
	curl -L -s https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 -o $GOPATH/bin/dep
	chmod +x $GOPATH/bin/dep
	dep ensure

clean:
	rm -rf $(PACKAGE) $(EXEC)
