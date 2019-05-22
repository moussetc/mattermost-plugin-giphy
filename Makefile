#
# Makefile for Mattermost Giphy Plugin
#
SRC=plugin.go post.go shuffle.go giphy.go gfycat.go
EXEC=plugin
CONF=plugin.yaml
PACKAGE_BASENAME=mattermost-plugin-giphy
TEST=plugin_test.go giphy_test.go

all: test-coverage dist-all

build: vendor $(SRC) $(CONF)
	go build -o $(EXEC) $(SRC)

rebuild: clean build

$(PACKAGE_BASENAME).tar.gz: build
	chmod a+x $(EXEC)
	tar -czvf $@ $(EXEC) $(CONF)

TAR_PLUGIN_EXE_TRANSFORM = --transform 'flags=r;s|dist/intermediate/plugin_.*|$(EXEC)|'
ifneq (,$(findstring bsdtar,$(shell tar --version)))
	TAR_PLUGIN_EXE_TRANSFORM = -s '|dist/intermediate/plugin_.*|$(EXEC)|'
endif

dist-all: vendor $(SRC) $(CONF)
	rm -rf ./dist
	go get github.com/mitchellh/gox
	$(shell go env GOPATH)/bin/gox -osarch='darwin/amd64 linux/amd64 windows/amd64' -output 'dist/intermediate/plugin_{{.OS}}_{{.Arch}}'
	tar -czvf dist/$(PACKAGE_BASENAME)-darwin-amd64.tar.gz $(TAR_PLUGIN_EXE_TRANSFORM) dist/intermediate/plugin_darwin_amd64 $(CONF)
	tar -czvf dist/$(PACKAGE_BASENAME)-linux-amd64.tar.gz $(TAR_PLUGIN_EXE_TRANSFORM) dist/intermediate/plugin_linux_amd64 $(CONF)
	tar -czvf dist/$(PACKAGE_BASENAME)-windows-amd64.tar.gz $(TAR_PLUGIN_EXE_TRANSFORM) dist/intermediate/plugin_windows_amd64.exe $(CONF)
	rm -rf dist/intermediate

test: $(SRC) $(TEST)
	go test -v .

test-coverage: $(SRC) $(TEST)
	go test -race -coverprofile=coverage.txt -covermode=atomic

vendor: Gopkg.lock
	go get github.com/golang/dep
	dep ensure

clean:
	rm -rf ./dist $(EXEC) $(PACKAGE_BASENAME).tar.gz
