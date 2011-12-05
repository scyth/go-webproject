
DIRS=\
	sessions\


install.dirs: $(addsuffix .install, $(DIRS))
clean.dirs: $(addsuffix .clean, $(DIRS))

%.install:
	$(MAKE) -C $*

%.clean:
	$(MAKE) -C $* clean
	
install: install.dirs
clean: clean.dirs
	
