GWPROOT := $(shell pwd)
export
include ./Make.inc

all:
	$(MAKE) -C src/sys/
	$(MAKE) -C src/modules/
	
	$(EXT)g -o build/$(TARGET).$(EXT) -I $(INCPATH) src/sys/main.go src/handlers.go
	$(EXT)l -o $(TARGET) -L $(INCPATH) build/$(TARGET).$(EXT)	

clean:
	$(MAKE) -C src/sys/ clean
	$(MAKE) -C src/modules/ clean
	rm -f build/$(TARGET).$(EXT) $(TARGET)
	rm -f build/gwp/*.a
	find ./ -name "*~" | xargs rm -f
