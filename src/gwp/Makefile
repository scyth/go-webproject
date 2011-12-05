include $(GWPROOT)/Make.inc

DIRS=\
	libs/goconf\
	libs/gorilla/context\
	libs/gorilla/mux\


install.dirs: $(addsuffix .install, $(DIRS))
clean.dirs: $(addsuffix .clean, $(DIRS))

%.install:
	# build supporting libs
	$(MAKE) -C $*

	# build gwp_context
	$(EXT)g -o gwp_context/_go.$(EXT) -I $(INCPATH) gwp_context/gwp_context.go gwp_context/doc.go
	gopack grc $(SYSPATH)/gwp_context.a gwp_context/_go.$(EXT)

	# build template API
	$(EXT)g -o gwp_template/_go.$(EXT) -I $(INCPATH) gwp_template/gwp_template.go gwp_template/doc.go
	gopack grc $(SYSPATH)/gwp_template.a gwp_template/_go.$(EXT)
	
	# build gwp_core
	$(EXT)g -o gwp_core/_go.$(EXT) -I $(INCPATH) gwp_core/gwp_core.go gwp_core/doc.go
	gopack grc $(SYSPATH)/gwp_core.a gwp_core/_go.$(EXT)
		
%.clean:
	$(MAKE) -C $* clean
	find ./ -name "_go.$(EXT)" | xargs rm -f 
	
install: install.dirs
clean: clean.dirs
