# Author: Paul Côté
# Last Change Author: Paul Côté
# Last Date Changed: 2022/06/10

PKG := github.com/SSSOC-CAN/laniakea

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v

# ============
# INSTALLATION
# ============

build:
	$(GOBUILD) -tags="${tags}" -o ./build/${bin} $(PKG)/cmd/laniakea
	$(GOBUILD) -tags="${tags}" -o ./build/${bincli} $(PKG)/cmd/lanicli

install:
	$(GOINSTALL) -tags="${tags}" $(PKG)/cmd/laniakea
	$(GOINSTALL) -tags="${tags}" $(PKG)/cmd/lanicli