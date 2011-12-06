include $(GWPROOT)/Make.inc

all:
	mkdir -p $(SYSPATH)/libs/gorilla
	$(EXT)g -o _go_.$(EXT) -I $(INCPATH) mux.go doc.go
	gopack grc $(SYSPATH)/libs/gorilla/mux.a _go_.$(EXT)

clean:
	
	rm -f $(SYSPATH)/libs/gorilla/mux.a
	rm -f _go_.$(EXT)
		
	