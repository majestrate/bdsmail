GOPATH=$(PWD)
LUA_SRC=$(GOPATH)/src/bds/lib/lua/lua/src
LUA_LIB=$(LUA_SRC)/libluajit.a


all: $(LUA_LIB) go

$(LUA_LIB):
	$(MAKE) -C $(LUA_SRC) $(MOPTS)

go: $(LUA_LIB)
	go install -v bds/cmd/maild
	go install -v bds/cmd/bdsconfig

clean:
	make -C $(LUA_SRC) clean
	go clean -v
	rm -rf pkg
