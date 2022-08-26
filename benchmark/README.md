# Benchmark

## Benchmark Oracle Script Spec

This oracle script will act as a proxy. We have to send scenario number and value with it to execute the specified type of code.

```
Input {
    data_source_id: u64,
    scenario: u64,
    value: u64,
}

Output {
    dummy: u64,
}
```

### Prepare scenarios

- Scenario 1: ask_external_data
  - Value = Number of calling

### Execute scenarios

- Scenario 101: infinite_loop
  - Value = Do nothing
- Scenario 102: arithmatic_ops
  - Value = Number of loop
- Scenario 103: allocate_mem
  - Value = Size of memory
- Scenario 104: find_median
  - Value = Number of loop

## How to run benchmark

```
go test -v -bench=. -benchtime=1s -benchmem -cpu 4
```
