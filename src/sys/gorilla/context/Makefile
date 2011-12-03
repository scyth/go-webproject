include ../../../../Make.inc

ARCH := $(shell uname -m)
ARCH_64 := x86_64

ifeq ($(ARCH), $(ARCH_64))
	EXT := 6
else
	EXT := 8
endif


all:
	mkdir -p $(SYSPATH)/gorilla
	$(EXT)g -o _go_.6 context.go doc.go
	gopack grc $(SYSPATH)/gorilla/context.a _go_.$(EXT)

clean:
	rm -f $(SYSPATH)/gorilla/context.a
	rm -f _go_.$(EXT)
