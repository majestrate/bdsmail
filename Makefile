LUA_SRC=$(PWD)/src/bds/lib/lua/lua/src
LUA_LIB=$(LUA_SRC)/libluajit.a


all: $(LUA_LIB) build

$(LUA_LIB):
	$(MAKE) -C $(LUA_SRC) $(MOPTS)

build: $(LUA_LIB)
	GOPATH=$(PWD) go install -v bds/cmd/maild
	GOPATH=$(PWD) go install -v bds/cmd/newmail
	GOPATH=$(PWD) go install -v bds/cmd/bdsconfig

clean:
	$(MAKE) -C $(LUA_SRC) clean
	go clean -v
	rm -rf pkg

test:
	go test bds/lib/...
