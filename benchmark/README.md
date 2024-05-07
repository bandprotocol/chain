# Benchmark

## Benchmark Oracle Script Spec

This oracle script will act as a proxy. We have to send the scenario number and value with it to execute the specified type of code.

```
Input {
    data_source_id: u64,
    scenario: u64,
    value: u64,
    text: string,
}

Output {
    dummy: u64,
}
```

### Prepare scenarios

- Scenario 1: ask_external_data
  - Value = Number of calling
- Scenario 2: infinite_loop
  - Value = -
- Scenario 3: arithmetic_ops
  - Value = Number of calling
- Scenario 4: allocate_mem
  - Value = Size of memory
- Scenario 5: find_median
  - Value = Number of calling
- Scenario 6: finite_loop
  - Value = Number of calling
- Scenario 7: set_local_var
  - Value = Size of memory
### Execute scenarios

- Scenario 0: Nothing
  - Value = -
- Scenario 101: infinite_loop
  - Value = -
- Scenario 102: arithmetic_ops
  - Value = Number of loops
- Scenario 103: allocate_mem
  - Value = Size of memory
- Scenario 104: find_median
  - Value = Number of loops
- Scenario 105: finite_loop
  - Value = Number of loops
- Scenario 106: set_local_var
  - Value = Size of memory
- Scenario 201: get_ask_count
  - Value = Number of loops
- Scenario 202: get_min_count
  - Value = Number of loops
- Scenario 203: get_prepare_time
  - Value = Number of loops
- Scenario 204: get_execute_time
  - Value = Number of loops
- Scenario 205: get_ans_count
  - Value = Number of loops
- Scenario 206: get_calldata
  - Value = Number of loops
- Scenario 207: save_return_data
  - Value = Number of loops
- Scenario 208: get_external_data
  - Value = Number of loops
- Scenario 209: ecvrf_verify
  - Value = Number of loops
- Scenario 210: base_import
  - Value = -

## How to run the benchmark

```
cd benchmark
go test -v -bench=. -benchtime=1s -benchmem -cpu 4 -parallel 1 -timeout 0
go test -v -bench=. -benchtime=5x -benchmem -cpu 4 -parallel 1 -timeout 0
```
