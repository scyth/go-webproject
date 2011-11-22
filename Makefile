all:
	$(MAKE)	-C goconf/
	$(MAKE) -C src/

clean:
	rm -f ./runserver *~ ./config/*~ ./src/*~ ./src/_go_.* ./templates/*~
	$(MAKE) -C goconf/ clean
