cmds=$(notdir $(wildcard cmd/*))
pkgs=$(wildcard internal/*)

all: fmt_all build_all

fmt_all:
	$(foreach c,$(cmds),go fmt cmd/$c/*.go; )
	$(foreach p,$(pkgs),go fmt $p/*.go; )

build_all: $(cmds)

define build_cmd =
$1: cmd/$1/$1.go
	go build cmd/$1/$1.go
endef

$(foreach c,$(cmds),$(eval $(call build_cmd,$c)))

clean:
	rm -f $(cmds)

.PHONY: clean fmt_all build_all
