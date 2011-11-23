Go-WebProject
=============

A lightweight toolkit/framework to get you started with web programming in Go.
It compiles with the latest weekly release of Go, so keep your code up2date! I tend to keep it this way until Go v1 comes out in a month or two.


### Features

* text/template package from standard go library is used
* template caching is in place
* (optional) integrated mux and context packages from gorilla project for advanced routing and easier request handling
* external configuration file for server and app settings


### References

* Check out http://code.google.com/p/gorilla/ for mux and context documentation


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

Enjoy!
