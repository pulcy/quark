PROJECT := droplets
SCRIPTDIR := $(shell pwd)
ROOTDIR := $(shell cd $(SCRIPTDIR) && pwd)
VERSION:= $(shell cat $(ROOTDIR)/VERSION)
COMMIT := $(shell git rev-parse --short HEAD)

GOBUILDDIR := $(SCRIPTDIR)/.gobuild
SRCDIR := $(SCRIPTDIR)
BINDIR := $(ROOTDIR)/bin

ORGPATH := arvika.pulcy.com/iggi
ORGDIR := $(GOBUILDDIR)/src/$(ORGPATH)
REPONAME := $(PROJECT)
REPODIR := $(ORGDIR)/$(REPONAME)
REPOPATH := $(ORGPATH)/$(REPONAME)
BIN := $(BINDIR)/$(PROJECT)

GOPATH := $(GOBUILDDIR)

ifndef GOOS
	GOOS := $(shell go env GOOS)
endif
ifndef GOARCH
	GOARCH := $(shell go env GOARCH)
endif	

SOURCES := $(shell find $(SRCDIR) -name '*.go')

.PHONY: all clean deps 

all: $(BIN)

clean:
	rm -Rf $(BIN) $(GOBUILDDIR)

deps: 
	@${MAKE} -B -s $(GOBUILDDIR)

$(GOBUILDDIR): 
	@mkdir -p $(ORGDIR)
	@rm -f $(REPODIR) && ln -s ../../../.. $(REPODIR)
	cd $(GOPATH) && devtool go get github.com/spf13/pflag
	cd $(GOPATH) && devtool go get github.com/spf13/cobra
	cd $(GOPATH) && devtool go get github.com/digitalocean/godo
	cd $(GOPATH) && devtool go get code.google.com/p/goauth2/oauth
	cd $(GOPATH) && devtool go get github.com/ryanuber/columnize
	cd $(GOPATH) && devtool go get github.com/dchest/uniuri
	
$(BIN): $(GOBUILDDIR) $(SOURCES) 
	@mkdir -p $(BINDIR)
	docker run \
	    --rm \
	    -v $(ROOTDIR):/usr/code \
	    -e GOPATH=/usr/code/.gobuild \
	    -e GOOS=$(GOOS) \
	    -e GOARCH=$(GOARCH) \
	    -w /usr/code/ \
	    golang:1.5.1 \
	    go build -a -ldflags "-X main.projectVersion=$(VERSION) -X main.projectBuild=$(COMMIT)" -o /usr/code/$(PROJECT)

