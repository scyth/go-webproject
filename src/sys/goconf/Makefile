include ../../../Make.inc

all:
	mkdir -p $(SYSPATH)
	$(EXT)g -o _go_.$(EXT) -I $(INCPATH) conf.go get.go read.go write.go
	gopack grc $(SYSPATH)/goconf.a _go_.$(EXT)
	
clean:
	rm -f $(SYSPATH)/goconf.a
	rm -f _go_.$(EXT)
