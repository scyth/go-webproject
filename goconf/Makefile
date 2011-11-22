
ARCH := $(shell uname -m)
ARCH_64 := x86_64

ifeq ($(ARCH), $(ARCH_64))
	EXT := 6
else
	EXT := 8
endif


all:
	$(EXT)g -o goconf.$(EXT) conf.go get.go read.go write.go

clean:
	
	rm -f goconf.$(EXT) *~
	
	