PKG := github.com/SSSOC-CAN/fmtd

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v

# ============
# INSTALLATION
# ============

build:
	$(GOBUILD) $(PKG)/cmd/fmtd
	$(GOBUILD) $(PKG)/cmd/fmtcli

install:
	$(GOINSTALL) $(PKG)/cmd/fmtd
	$(GOINSTALL) $(PKG)/cmd/fmtcli