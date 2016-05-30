# BDS Mail #

Brain Dead Simple Mail Server

Server that runs i2p.rocks email


![logo](https://github.com/majestrate/bdsmail/raw/master/contrib/assets/logo.png "logo")

## Usage ##

To use this software you need to own a domain.

If you're lazy and/or cheap sign up for a free dynamic dns with
[no-ip](https://www.noip.com/) or a related provider.


### Building From Source ###

The only build dependancy is go 1.6

    # apt install golang-1.6

To build the daemon do:

    $ git clone https://github.com/majestrate/bdsmail
    $ cd bdsmail
    $ ./build.sh

This will yield 2 executables `bdsmail` and `bdsconfig`


### Configuring ###


To generate an initial configuration file run the following:

    $ ./bdsconfig yourdomain.tld > bdsmail.lua

### Running ###

Copy `bdsmail` and `bdsmail.lua` to the machine that `yourdomain.tld` points to.

Then on that machine run the mail server:

    $ ./bdsmail bdsmail.lua

Finnally, forward port 25 to 2525 so that inbound mail reaches the mail server (requires root)

    # iptables -t nat -A PREROUTING -i eth0 -p tcp --dport 25 -j REDIRECT --to-port 2525

If everything went smoothly, anyone on the internet can email `admin@yourdomain.tld` now.

## Features ##

### Current ###

* brain dead simple configuration (see config.example.lua)
* brain dead simple database backend (sqlite3)
* brain dead simple license (MIT)

### Near Future ###

* brain dead simple tls enabled by default
* brain dead simple smtp access
* brain dead simple pop3 access

### Future (Eventually) ###

* brain dead simple i2pbote gateway
* brain dead simple inet/i2p mail relay
