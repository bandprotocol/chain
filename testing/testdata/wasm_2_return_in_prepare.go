package testdata

// Wasm2 is a bad Owasm script with the following specification:
//
//	PREPARE:
//	  CALL set_return_data with RETDATA "test" -- Not allowed during prepare
//	EXECUTE:
//	  DO NOTHING
var Wasm2 []byte = wat2wasm(`
(module
	(type $t0 (func))
	(type $t2 (func (param i64 i64)))
	(import "env" "set_return_data" (func $set_return_data (type $t2)))
	(func $prepare (export "prepare")
		i64.const 1024
		i64.const 4
		call $set_return_data)
	(func $execute (export "execute"))
	(memory $memory (export "memory") 17)
	(data (i32.const 1024) "test"))
`)
