
libs=\
	github.com/scyth/go-webproject/gwp/libs/goconf\
	github.com/scyth/go-webproject/gwp/libs/gorilla/context\
	github.com/scyth/go-webproject/gwp/libs/gorilla/mux/\


packages=\
	github.com/scyth/go-webproject/gwp/gwp_core\
	github.com/scyth/go-webproject/gwp/gwp_context\
	github.com/scyth/go-webproject/gwp/gwp_template\
	github.com/scyth/go-webproject/gwp/gwp_module\


modules=\
	github.com/scyth/go-webproject/gwp/modules/mod_sessions\
	github.com/scyth/go-webproject/gwp/modules/mod_example\

all:
	$(MAKE) install
	

sync:
	# Downloading packages
	$(MAKE) sync.dirs


install:
	# Installing packages
	$(MAKE) install.dirs
	# Installing binary
	go build -o runserver src/main.go src/handlers.go
	

clean:
	# Cleaning up packages
	$(MAKE) clean.dirs
	# Cleaning up source
	find ./ -name "*~" | xargs rm -f
	# Cleaning up binary
	rm -f ./runserver


reinstall:
	$(MAKE) clean
	$(MAKE) install
	

sync.dirs: $(addsuffix .sync, $(libs) $(packages) $(modules))
install.dirs: $(addsuffix .install, $(libs) $(packages) $(modules))
clean.dirs: $(addsuffix .clean, $(libs) $(packages) $(modules))

	
%.sync:
	go get -d $*

	
%.install:
	go install $*


%.clean:
	go clean $*


