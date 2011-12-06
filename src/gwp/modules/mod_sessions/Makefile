include $(GWPROOT)/Make.inc

all:
	$(EXT)g -o _go_.$(EXT) -I $(INCPATH) sessions.go sessions_addon.go filestore.go mod_sessions.go doc.go
	gopack grc $(MODPATH)/mod_sessions.a _go_.$(EXT)

clean:
	
	rm -f $(MODPATH)/mod_sessions.a
	rm -f _go_.$(EXT)
