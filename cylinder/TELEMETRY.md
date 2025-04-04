# Cylinder Telemetry

## Metrics Tracked

The following telemetry metrics are captured and exposed via Prometheus:

### DKG Process Metrics

#### Round 1

- `process_round1_success_count` (Counter): Number of successful process round 1
  - Labels: `group_id`
- `process_round1_failure_count` (Counter): Number of failed process round 1
  - Labels: `group_id`
- `process_round1_time` (Summary): Time taken to process round 1
  - Labels: `group_id`
  - Percentiles: 50th, 90th, 99th

#### Round 2

- `process_round2_success_count` (Counter): Number of successful process round 2
  - Labels: `group_id`
- `process_round2_failure_count` (Counter): Number of failed process round 2
  - Labels: `group_id`
- `process_round2_time` (Summary): Time taken to process round 2
  - Labels: `group_id`
  - Percentiles: 50th, 90th, 99th

#### Round 3

- `process_round3_confirm_count` (Counter): Number of successful process round 3 confirm
  - Labels: `group_id`
- `process_round3_complain_count` (Counter): Number of process round 3 complain
  - Labels: `group_id`
- `process_round3_failure_count` (Counter): Number of failed process round 3
  - Labels: `group_id`
- `process_round3_time` (Summary): Time taken to process round 3
  - Labels: `group_id`
  - Percentiles: 50th, 90th, 99th

### Group and DKG Metrics

- `dkg_left_gauge` (Gauge): Number of DKG left in the store
- `group_count` (Counter): Number of groups in the store

### DE (Data Enclave) Metrics

- `on_chain_de_left_gauge` (Gauge): Number of on-chain DE left
- `off_chain_de_left_gauge` (Gauge): Number of DE left in the store
- `de_count_used_gauge` (Gauge): Number of DE count used

### Signing Metrics

- `incoming_signing_count` (Counter): Number of incoming signing requests
  - Labels: `group_id`
- `process_signing_success_count` (Counter): Number of successful process signing
  - Labels: `group_id`
- `process_signing_failure_count` (Counter): Number of failed process signing
  - Labels: `group_id`
- `process_signing_time` (Summary): Time taken to process signing
  - Labels: `group_id`
  - Percentiles: 50th, 90th, 99th

### Transaction Metrics

- `waiting_sender_time` (Summary): Time taken to wait for a free key
  - Percentiles: 50th, 90th, 99th
- `submitting_tx_count` (Counter): Number of submitting transactions
- `submit_tx_success_count` (Counter): Number of successful submit transactions
- `submit_tx_failed_count` (Counter): Number of failed submit transactions
- `submit_tx_time` (Summary): Time taken to submit transactions
  - Percentiles: 50th, 90th, 99th

## Grafana Dashboard

Cylinder provides a pre-built Grafana Dashboard to visualize TSS signing metrics efficiently. You can download and import the dashboard from Grafana's official repository.

### Dashboard Link

[Cylinder Grafana Dashboard #23071](https://grafana.com/grafana/dashboards/23184-cylinder/)

## Metrics Server Configuration

The metrics server can be configured using:

- Configuration file: `metrics-listen-addr` setting
- Command line: `--metrics-listen-addr` flag

The server exposes metrics at the `/metrics` endpoint in Prometheus format.
