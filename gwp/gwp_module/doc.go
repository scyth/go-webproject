/*
Package gwp_module gives API for writing 3rd party modules.

Modules can extend functionality in variety of ways:

* every module gets a chance to export its own custom parameters to global server.conf, 
which makes it perfect for integration of 3rd party Go packages, like database libs. 
See mod_sessions as an example of such a module.

* every module gets limited access to internal calls for managing runtime. API is exposed through this package.

* a module can register handlers, which can make it sort of sub-application that can be reused by many.
mod_example provided, ilustrates this.

* a module can spawn goroutines for background tasks that are not request based.

*/
package gwp_module
