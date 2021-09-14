PKG := github.com/SSSOC-CAN/fmtd

GOBUILD := GO111MODULE=on go build -v
GOINSTALL := GO111MODULE=on go install -v

# ============
# INSTALLATION
# ============

build:
	$(GOBUILD) -o fmtd-debug $(PKG)/cmd/fmtd
	$(GOBUILD) -o fmtcli-debug $(PKG)/cmd/fmtcli

install:
	$(GOINSTALL) $(PKG)/cmd/fmtd
	$(GOINSTALL) $(PKG)/cmd/fmtcli