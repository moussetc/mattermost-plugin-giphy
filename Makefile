#
# Makefile for Mattermost Giphy Plugin
#

# REMEMBER to set the OS and ARCHITECTURE of the
# Mattermost server for the Go cross-compilation :
TARGET_OS=linux
TARGET_ARCH=amd64

DEP_VERSION=0.4.1

SRC=plugin.go gifProvider.go
EXEC=plugin.exe
CONF=plugin.yaml
PACKAGE_BASENAME=mattermost-giphy-plugin
TEST=plugin_test.go gifProvider.go

all: test-coverage dist

$(EXEC): $(SRC)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -o $(EXEC) $(SRC)

build: $(EXEC)

rebuild: clean build

TAR_PLUGIN_EXE_TRANSFORM = --transform 'flags=r;s|dist/intermediate/plugin_.*|plugin.exe|'
ifneq (,$(findstring bsdtar,$(shell tar --version)))
	TAR_PLUGIN_EXE_TRANSFORM = -s '|dist/intermediate/plugin_.*|plugin.exe|'
endif

dist: vendor $(EXEC) $(CONF)
	rm -rf ./dist
	go get github.com/mitchellh/gox
	$(shell go env GOPATH)/bin/gox -osarch='darwin/amd64 linux/amd64 windows/amd64' -output 'dist/intermediate/plugin_{{.OS}}_{{.Arch}}'
	tar -czvf dist/$(PACKAGE_BASENAME)-darwin-amd64.tar.gz $(TAR_PLUGIN_EXE_TRANSFORM) dist/intermediate/plugin_darwin_amd64 plugin.yaml
	tar -czvf dist/$(PACKAGE_BASENAME)-linux-amd64.tar.gz $(TAR_PLUGIN_EXE_TRANSFORM) dist/intermediate/plugin_linux_amd64 plugin.yaml
	tar -czvf dist/$(PACKAGE_BASENAME)-windows-amd64.tar.gz $(TAR_PLUGIN_EXE_TRANSFORM) dist/intermediate/plugin_windows_amd64.exe plugin.yaml
	rm -rf dist/intermediate

test: $(SRC) $(TEST)
	go test -v .

test-coverage: $(SRC) $(TEST)
	go test -race -coverprofile=coverage.txt -covermode=atomic

vendor: Gopkg.lock
	curl -L -s https://github.com/golang/dep/releases/download/v${DEP_VERSION}/dep-linux-amd64 -o $GOPATH/bin/dep
	chmod +x $GOPATH/bin/dep
	dep ensure

clean:
	rm -rf ./dist $(EXEC)
