include $(GWPROOT)/Make.inc

all:
	mkdir -p $(SYSPATH)/libs/gorilla
	go tool $(EXT)g -o _go_.$(EXT) -I $(INCPATH) mux.go doc.go
	go tool pack grc $(SYSPATH)/libs/gorilla/mux.a _go_.$(EXT)

clean:
	
	rm -f $(SYSPATH)/libs/gorilla/mux.a
	rm -f _go_.$(EXT)
		
	