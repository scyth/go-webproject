all:
	$(MAKE) -C gorilla/context/
	$(MAKE) -C gorilla/mux/
	$(MAKE)	-C goconf/
	$(MAKE) -C src/

clean:
	rm -f ./runserver
	find ./ -name "*~" | xargs rm -f
	find ./ -name "_go_.*" | xargs rm -f
	$(MAKE) -C gorilla/context/ clean
	$(MAKE) -C gorilla/mux/ clean
	$(MAKE) -C goconf/ clean

