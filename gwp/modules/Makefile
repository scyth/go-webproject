
DIRS=\
	mod_sessions\
	mod_example\


install.dirs: $(addsuffix .install, $(DIRS))
clean.dirs: $(addsuffix .clean, $(DIRS))

%.install:
	mkdir -p $(MODPATH)
	$(MAKE) -C $*

%.clean:
	$(MAKE) -C $* clean
	
install: install.dirs
clean: clean.dirs
	
