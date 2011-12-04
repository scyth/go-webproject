include ../../Make.inc

DIRS=\
	goconf\
	gorilla/context\
	gorilla/mux\


install.dirs: $(addsuffix .install, $(DIRS))
clean.dirs: $(addsuffix .clean, $(DIRS))

%.install:
	# build from subdirectories
	$(MAKE) -C $*

	# build gwp_context
	$(EXT)g -o _gwp_context.$(EXT) -I $(INCPATH) gwp_context.go
	gopack grc $(SYSPATH)/gwp_context.a _gwp_context.$(EXT)

	# build template API
	$(EXT)g -o _gwp_template.$(EXT) -I $(INCPATH) gwp_template.go
	gopack grc $(SYSPATH)/gwp_template.a _gwp_template.$(EXT)
	
	# build gwp_core
	$(EXT)g -o _gwp_core.$(EXT) -I $(INCPATH) gwp_core.go
	gopack grc $(SYSPATH)/gwp_core.a _gwp_core.$(EXT)
		
%.clean:
	$(MAKE) -C $* clean
	rm -f ./_*.$(EXT)
	
install: install.dirs
clean: clean.dirs
