GOPATH=$(PWD)
LUA_SRC=$(GOPATH)/src/bds/lib/lua/lua/src
LUA_LIB=$(LUA_SRC)/libluajit.a


all: $(LUA_LIB) build

$(LUA_LIB):
	$(MAKE) -C $(LUA_SRC) $(MOPTS)

build: $(LUA_LIB)
	go install -v bds/cmd/maild
	go install -v bds/cmd/newmail
	go install -v bds/cmd/bdsconfig

clean:
	make -C $(LUA_SRC) clean
	go clean -v
	rm -rf pkg
