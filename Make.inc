
SYSPATH := $(GWPROOT)/build/gwp
MODPATH := $(GWPROOT)/build/gwp/modules
INCPATH := $(GWPROOT)/build
TARGET  := runserver

ARCH := $(shell uname -m)
ARCH_64 := x86_64

ifeq ($(ARCH), $(ARCH_64))
        EXT := 6
else
        EXT := 8
endif
