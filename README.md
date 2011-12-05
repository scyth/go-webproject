Go-WebProject
=============

A lightweight toolkit/framework to get you started with web programming in Go.
It compiles with the latest weekly release of Go(weekly.2011-12-02), so keep your code up2date! I tend to keep it this way until Go v1 comes out in a month or two.


### Features

* html/template package from standard go library is used, for consistency and security reasons.
* template caching is in place.
* (optional) on the fly template reloading without need to restart the service.
* (optional) integrated mux package from gorilla project for advanced routing and easier request handling.
* integrated sessions package from gorilla project for session management.
* external configuration file for server and app settings.


### References

* Check out http://go-webproject.appspot.com/ for more documentation about internal packages


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

Enjoy!
