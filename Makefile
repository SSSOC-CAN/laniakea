# Author: Paul Côté
# Last Change Author: Paul Côté
# Last Date Changed: 2022/06/10

PKG := github.com/SSSOC-CAN/fmtd

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v

# ============
# INSTALLATION
# ============

build:
	$(GOBUILD) -tags="${tags}" -o fmtd-debug $(PKG)/cmd/fmtd
	$(GOBUILD) -tags="${tags}" -o fmtcli-debug $(PKG)/cmd/fmtcli

install:
	$(GOINSTALL) -tags="${tags}" $(PKG)/cmd/fmtd
	$(GOINSTALL) -tags="${tags}" $(PKG)/cmd/fmtcli