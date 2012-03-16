include $(GWPROOT)/Make.inc

install:
	# build supporting libs
	$(MAKE) -C libs/goconf
	$(MAKE) -C libs/gorilla/context
	$(MAKE) -C libs/gorilla/mux

	# build gwp_context
	go tool $(EXT)g -o gwp_context/_go.$(EXT) -I $(INCPATH) gwp_context/gwp_context.go gwp_context/doc.go
	go tool pack grc $(SYSPATH)/gwp_context.a gwp_context/_go.$(EXT)

	# build template API
	go tool $(EXT)g -o gwp_template/_go.$(EXT) -I $(INCPATH) gwp_template/gwp_template.go gwp_template/doc.go
	go tool pack grc $(SYSPATH)/gwp_template.a gwp_template/_go.$(EXT)

	# build gwp_core
	go tool $(EXT)g -o gwp_core/_go.$(EXT) -I $(INCPATH) gwp_core/gwp_core.go gwp_core/doc.go
	go tool pack grc $(SYSPATH)/gwp_core.a gwp_core/_go.$(EXT)

	# build gwp_module
	go tool $(EXT)g -o gwp_module/_go.$(EXT) -I $(INCPATH) gwp_module/gwp_module.go gwp_module/doc.go
	go tool pack grc $(SYSPATH)/gwp_module.a gwp_module/_go.$(EXT)	

clean:
	$(MAKE) -C libs/goconf clean
	$(MAKE) -C libs/gorilla/context clean
	$(MAKE) -C libs/gorilla/mux clean
	find ./ -name "_go.$(EXT)" | xargs rm -f 
	
all: install
