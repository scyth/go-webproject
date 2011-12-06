include $(GWPROOT)/Make.inc

all:
	$(EXT)g -o _go_.$(EXT) -I $(INCPATH) mod_example.go
	gopack grc $(MODPATH)/mod_example.a _go_.$(EXT)

clean:
	
	rm -f $(MODPATH)/mod_example.a
	rm -f _go_.$(EXT)
