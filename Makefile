GOBIN = $(firstword $(subst :, ,$(GOPATH)))/bin

all: test

.PHONY: test
test: $(GOBIN)/goconvey
	go test -v

doc: README.md

README.md: $(GOBIN)/godocdown *.go
	$< > $@

$(GOBIN)/godocdown:
	go get github.com/robertkrimen/godocdown/godocdown

$(GOBIN)/goconvey:
	go get github.com/smartystreets/goconvey/convey
