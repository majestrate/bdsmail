package lua

// #cgo CFLAGS: -I ${SRCDIR}/lua/src
// #cgo LDFLAGS: ${SRCDIR}/lua/src/libluajit.so -lm -ldl
// #include <luajit.h>
// #include <lualib.h>
// #include <lauxlib.h>
// #include <stdlib.h>
// #include <string.h>
import "C"

import (
	"errors"
	log "github.com/Sirupsen/logrus"
	"unsafe"
)

var ErrNoJIT = errors.New("jit not enabled")

const configvar = "config"

// lua interpreter
type Lua struct {
	state *C.lua_State
}

// do full GC cycle
func (l *Lua) GC() {
	C.lua_gc(l.state, C.LUA_GCCOLLECT, 0)
}

// get a configuration option
// return the value and true if found
// return emtpy string and false if not found
func (l *Lua) GetConfigOpt(name string) (val string, ok bool) {
	log.Debug("get config opt ", name)
	cname := C.CString(name)
	C.lua_getfield(l.state, C.LUA_GLOBALSINDEX, cname)
	// ... as a string (this value is freed by lua)
	var clen C.size_t
	cstr := C.lua_tolstring(l.state, -1, &clen)
	// convert to a buffer we own
	if clen > 0 {
		b := C.GoBytes(unsafe.Pointer(cstr), C.int(clen))
		v := make([]byte, len(b))
		copy(v, b)
		val = string(v)
		ok = true
	}
	// free it
	C.free(unsafe.Pointer(cname))
	return
}

// call a lua function called `funcname` to filter a mail message
// return an integer that the function returns or -1 on error
// lua call is in the following format:
//
// res = funcname(addr, recip, sender, body)
//
func (l *Lua) CallMailFilter(funcname, addr, recip, sender, body string) (ret int) {
	cf := C.CString(funcname)
	C.lua_getfield(l.state, C.LUA_GLOBALSINDEX, cf)
	ca := C.CString(addr)
	C.lua_pushstring(l.state, ca)
	cr := C.CString(recip)
	C.lua_pushstring(l.state, cr)
	cs := C.CString(sender)
	C.lua_pushstring(l.state, cs)
	cb := C.CString(body)
	C.lua_pushstring(l.state, cb)
	C.lua_call(l.state, 4, 1)
	cret := C.lua_tointeger(l.state, -1)
	// convert return value to int
	ret = int(cret)
	// free buffers
	C.free(unsafe.Pointer(cb))
	C.free(unsafe.Pointer(cs))
	C.free(unsafe.Pointer(cr))
	C.free(unsafe.Pointer(ca))
	C.free(unsafe.Pointer(cf))
	// do lua gc
	l.GC()
	return
}

func (l *Lua) LoadFile(fname string) (err error) {
	cfname := C.CString(fname)
	defer C.free(unsafe.Pointer(cfname))
	res := C.luaL_loadfile(l.state, cfname)
	if res == 0 {
		log.Info("loaded config ", fname)
		res = C.lua_pcall(l.state, 0, 0, 0)
		if res != 0 {
			err = errors.New(C.GoString(C.lua_tolstring(l.state, -1, nil)))
		}
	} else {
		// failed to load file
		err = errors.New("failed to load file " + fname)
	}
	return
}

// close the interpreter
// all resources are expunged and no operations can be done after this
func (l *Lua) Close() {
	if l.state != nil {
		C.lua_close(l.state)
	}
	l.state = nil
}

// turn jit on globally
func (l *Lua) JIT() (err error) {
	flags := C.int(C.LUAJIT_MODE_ON)
	flags |= C.int(C.LUAJIT_MODE_ENGINE)
	res := C.luaJIT_setmode(l.state, 0, flags)
	if res == 0 {
		// failed to enable jit
		err = ErrNoJIT
	}
	return
}

// create a new lua interpreter
func New() (l *Lua) {
	l = new(Lua)
	l.state = C.luaL_newstate()
	if l.state == nil {
		l = nil
	} else {
		// open stdlib
		C.luaL_openlibs(l.state)
	}
	return
}
