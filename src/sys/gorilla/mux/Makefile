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
	$(EXT)g -o _go_.$(EXT) -I $(INCPATH) mux.go doc.go
	gopack grc $(SYSPATH)/gorilla/mux.a _go_.$(EXT)

clean:
	
	rm -f $(SYSPATH)/gorilla/mux.a
	rm -f _go_.$(EXT)
		
	