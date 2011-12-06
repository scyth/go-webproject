Go-WebProject
=============

Go-webproject is an opensource web application framework written in The Go Programming Language.

It is not a Go package that you can import elsewhere. It is completely standalone application which 
consists of a native, high performant web server and surrounding framework for application code. 
It is designed to be very easy to start coding with, but still easily extensible, with its flexible modular design. 
Project is actively maintained and it compiles with the latest Go weekly release - which is at the moment: weekly.2012-12-02.


### Features

* html/template package from standard go library is used, for consistency and security reasons.
* template caching is in place.
* (optional) on the fly template reloading without need to restart the service.
* (optional) integrated mux package from gorilla project for advanced routing and easier request handling.
* integrated sessions package from gorilla project for session management.
* extensions through 3rd party modules
* external configuration file for server and app settings.


### References

* Check out http://go-webproject.appspot.com/ for more documentation.
* or you can run ` godoc -http:8080 ./src/gwp ` to locally browse package documentation from your browser.


### Notes

* When live-templates feature is used, due to current bug in exp/inotify package, when template file is being Watch()ed, if it gets removed and 
replaced with another file of same name, updates to the new file are not tracked. I expect this to be resolved in future.


Installation
------------

### Step 1

* `$ git clone git://github.com/scyth/go-webproject.git`
* `$ cd go-webproject`


### Step 2

Copy the configuration file from examples/config/server.conf to config/server.conf and make appropriate changes for your environment.


### Step 3

Copy sample templates from example/templates/ to your templatePath directory you previously defined.


### Step 4

Start programming your handlers! Copy examples/handlers/handlers.go to src/, edit src/handlers.go and write some stuff based on examples provided.


### Step 5

Compile and run the server, asuming your current working directory is still project root

* `$ make`
* `$ ./runserver -config=config/server.conf`

Open up your browser and see if your code works as expected.


Writing 3rd party modules
-------------------------

See: http://go-webproject.appespot.com/pkg/gwp/gwp_module.html
