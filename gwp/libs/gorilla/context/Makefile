include $(GWPROOT)/Make.inc

ARCH := $(shell uname -m)
ARCH_64 := x86_64

ifeq ($(ARCH), $(ARCH_64))
	EXT := 6
else
	EXT := 8
endif


all:
	mkdir -p $(SYSPATH)/libs/gorilla
	go tool $(EXT)g -o _go_.6 context.go doc.go
	go tool pack grc $(SYSPATH)/libs/gorilla/context.a _go_.$(EXT)

clean:
	rm -f $(SYSPATH)/libs/gorilla/context.a
	rm -f _go_.$(EXT)
