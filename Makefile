# Setup name variables for the package/tool
NAME := pepper
PKG := github.com/genuinetools/$(NAME)

CGO_ENABLED := 0

# Set any default go build tags.
BUILDTAGS :=

include basic.mk

.PHONY: prebuild
prebuild:
