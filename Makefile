GWPROOT := $(shell pwd)
export
include ./Make.inc

all:
	$(MAKE) -C src/gwp/
	$(MAKE) -C src/gwp/modules/
	
	$(EXT)g -o build/$(TARGET).$(EXT) -I $(INCPATH) src/cmd/main.go src/handlers.go
	$(EXT)l -o $(TARGET) -L $(INCPATH) build/$(TARGET).$(EXT)	

clean:
	$(MAKE) -C src/gwp/ clean
	$(MAKE) -C src/gwp/modules/ clean
	rm -rf build/*
	rm -f ./runserver
	find ./ -name "*~" | xargs rm -f
