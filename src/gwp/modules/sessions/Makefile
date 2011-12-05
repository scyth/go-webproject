include $(GWPROOT)/Make.inc

all:
	mkdir -p $(MODPATH)/gorilla
	$(EXT)g -o _go_.$(EXT) -I $(INCPATH) sessions.go sessions_addon.go filestore.go doc.go
	gopack grc $(MODPATH)/gorilla/sessions.a _go_.$(EXT)

clean:
	
	rm -f $(MODPATH)/gorilla/sessions.a
	rm -f _go_.$(EXT)
