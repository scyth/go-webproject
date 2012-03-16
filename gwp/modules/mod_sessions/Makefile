include $(GWPROOT)/Make.inc

all:
	go tool $(EXT)g -o _go_.$(EXT) -I $(INCPATH) sessions.go sessions_addon.go filestore.go mod_sessions.go doc.go
	go tool pack grc $(MODPATH)/mod_sessions.a _go_.$(EXT)

clean:
	
	rm -f $(MODPATH)/mod_sessions.a
	rm -f _go_.$(EXT)
