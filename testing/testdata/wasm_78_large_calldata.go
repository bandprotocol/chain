package testdata

import (
	"fmt"
	"strings"
)

// Wasm78 is an oracle script to test large calldata
//
//	PREPARE:
//	  Ask external data with calldata "b"*time
//	EXECUTE:
//	  Return with data "b"*time
func Wasm78(time int) []byte {
	var b strings.Builder
	for idx := 0; idx < time; idx++ {
		b.Write([]byte("b"))
	}
	return wat2wasm(fmt.Sprintf(`(module
	 (type $t0 (func))
	 (type $t1 (func (param i64 i64 i64 i64)))
	 (type $t2 (func (param i64 i64)))
	 (import "env" "ask_external_data" (func $ask_external_data (type $t1)))
	 (import "env" "set_return_data" (func $set_return_data (type $t2)))
	 (func $prepare (export "prepare") (type $t0)
	   (local $l0 i64)
	   i64.const 1
	   i64.const 1
	   i32.const 1024
	   i64.extend_i32_u
	   local.tee $l0
	   i64.const %d
	   call $ask_external_data)
	 (func $execute (export "execute") (type $t0)
	   i32.const 1024
	   i64.extend_i32_u
	   i64.const %d
	   call $set_return_data)
	 (table $T0 1 1 funcref)
	 (memory $memory (export "memory") 200)
	 (data (i32.const 1024) "%s"))
	  `, time, time, b.String()))
}
