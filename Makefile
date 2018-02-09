#
# Makefile for Mattermost Giphy Plugin
#

# REMEMBER to set the OS and ARCHITECTURE of the
# Mattermost server for the Go cross-compilation :
TARGET_OS=linux
TARGET_ARCH=amd64

SRC=plugin.go gifProvider.go
EXEC=plugin
CONF=plugin.yaml
PACKAGE=plugin.tar.gz
TEST=plugin_test.go

$(EXEC): $(SRC)
	GOOS=$(TARGET_OS) GOARCH=$(TARGET_ARCH) go build -o $(EXEC) $(SRC)

build: $(EXEC)

rebuild: clean build

dist: $(EXEC) $(CONF)
	@echo "BEWARE, if this command is executed on Windows, the executable bit of the plugin executable will NOT be set correctly..."
	chmod a+x $(EXEC) && tar -czvf $(PACKAGE) $(EXEC) $(CONF)

test: $(SRC) $(TEST)
	go test .
clean:
	rm -rf $(PACKAGE) $(EXEC)
