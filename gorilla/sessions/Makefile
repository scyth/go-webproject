
ARCH := $(shell uname -m)
ARCH_64 := x86_64

ifeq ($(ARCH), $(ARCH_64))
	EXT := 6
else
	EXT := 8
endif


all:
	$(EXT)g -o sessions.$(EXT) sessions.go sessions_addon.go filestore.go doc.go

clean:
	
	rm -f sessions.$(EXT)
	
	