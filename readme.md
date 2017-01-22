![logo](logo.png "logo")

Brain Dead Simple Mail Server ᴮᴱᵀᴬ

### Building From Source ###

Dependencies:

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

### Email setup ###

See the example config for mutt [here](contrib/config/mutt/muttrc)

### Contact ###

test email address is `test@ivpxmoh2qzcmxbij3sxfnlsua6panhxke2b3bbhn4xxw7oacujdq.b32.i2p` 

if the server is down ding me on xmpp: `jeff@i2p.rocks`

## Features ##

### Current ###

* brain dead simple self hosted i2p mail
* brain dead simple database backend (sqlite3)
* brain dead simple smtp access
* brain dead simple pop3 access
* brain dead simple license (MIT)

### Near Future ###

* brain dead simple i2pbote gateway

### Future (Eventually) ###

* brain dead simple inet/i2p mail relay
