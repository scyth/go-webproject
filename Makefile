GWPROOT := $(shell pwd)
export
include ./Make.inc

all:
	$(MAKE) -C src/gwp/
	$(MAKE) -C src/gwp/modules/
	
	go tool $(EXT)g -o build/$(TARGET).$(EXT) -I $(INCPATH) src/cmd/main.go src/handlers.go
	go tool $(EXT)l -o $(TARGET) -L $(INCPATH) build/$(TARGET).$(EXT)	

clean:
	$(MAKE) -C src/gwp/ clean
	$(MAKE) -C src/gwp/modules/ clean
	rm -rf build/*
	rm -f ./runserver
	find ./ -name "*~" | xargs rm -f
