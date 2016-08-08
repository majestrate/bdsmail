![logo](logo.png "logo")

Brain Dead Simple Mail Server


## This software is still in heavy development ##

### Building From Source ###

Dependancies:

* gcc 4.8 or higher
* go 1.6 
* git
* make
* i2p with SAM enabled

To build the daemon do:

    $ git clone https://github.com/majestrate/bdsmail
    $ cd bdsmail
    $ make

### Configuring ###


To generate an initial configuration file run the following:

    $ ./bin/bdsconfig > config.lua

### Running ###

    $ ./bin/maild config.lua

## Features ##

### Current ###

* brain dead simple self hosted i2p mail
* brain dead simple license (MIT)

### Near Future ###

* brain dead simple database backend (sqlite3)
* brain dead simple tls enabled by default
* brain dead simple smtp access
* brain dead simple pop3 access

### Future (Eventually) ###

* brain dead simple i2pbote gateway
* brain dead simple inet/i2p mail relay
