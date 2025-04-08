package testdata

// WasmExtra1 is an extra Owasm code to test creating or editing oracle scripts.
var WasmExtra1 []byte = wat2wasm(`
(module
	(type $t0 (func))
	(type $t2 (func (param i64 i64)))
	(import "env" "set_return_data" (func $set_return_data (type $t2)))
	(func $prepare (export "prepare") (type $t0))
	(func $execute (export "execute") (type $t0))
	(memory $memory (export "memory") 17))

`)

// WasmExtra2 is another extra Owasm code to test creating or editing oracle scripts.
var WasmExtra2 []byte = wat2wasm(`
(module
	(type $t0 (func))
	(type $t1 (func (param i64 i64 i64 i64)))
	(type $t2 (func (param i64 i64)))
	(import "env" "ask_external_data" (func $ask_external_data (type $t1)))
	(import "env" "set_return_data" (func $set_return_data (type $t2)))
	(func $prepare (export "prepare") (type $t0))
	(func $execute (export "execute") (type $t0))
	(memory $memory (export "memory") 17))
`)
