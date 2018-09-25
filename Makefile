# Setup name variables for the package/tool
NAME := pepper
PKG := github.com/genuinetools/$(NAME)

CGO_ENABLED := 0

include basic.mk
