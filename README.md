Go-WebProject
=============

A lightweight toolkit/framework to get you started with web programming in Go.
It compiles with the latest weekly release of Go, so keep your code up2date! I tend to keep it this way until Go v1 comes out in a month or two.


--- Features

* text/template package from standard go library is used
* template caching is in place
* basic route handling provided by http package
* external configuration file for app settings


Installation
------------

### Step 1

* `$ git clone git@github.com:scyth/go-webproject.git`
* `$ cd go-webproject`


### Step 2

Create your configuration file, like example below:

[default]
listen		= 127.0.0.1:8000
projectroot	= /path/to/go-webproject
templatepath	= /path/to/go-webproject/templates
tmpdir		= /tmp

and save it in config/server.conf file.

### Step 3

Start programming your handlers! Edit src/handlers.go and write some stuff based on examples provided.

### Step 4

Run the server

* `$ make` -- asuming your current working directory is project root
* `$ ./runserver -config=config/server.conf`
* Open up your browser and see if your code works as expected.

