# Grogu Telemetry

## Metrics Tracked

The following telemetry metrics are captured and exposed via Prometheus:

### Updater Metrics

- `grogu_update_registry_count` (Counter): Number of times the registry is updated since last grogu restart
- `grogu_update_registry_failed_count` (Counter): Number of times the registry fail to update since last grogu restart
- `grogu_update_registry_success_count` (Counter): Number of times the registry successfully update since last grogu restart

### Signaller

- `grogu_validator_status` (Gauge): Validator status (1 = active, 0 = inactive)
- `grogu_processing_signal_count` (Counter): Number of times the signaler processes signal prices
- `grogu_process_signal_skipped_count` (Counter): Number of times the signaler's process is skipped
- `grogu_process_signal_failed_count` (Counter): Number of times the signaler failed to process signal prices
- `grogu_process_signal_success_count` (Counter): Number of times the signaler successfully process signal prices
- `grogu_query_signal_prices_duration` (Summary): Time being consumed for querying signal prices from Bothan service
  - Percentiles: 50th, 90th, 99th
- `grogu_non_pending_signal_ids` (Gauge): Number of non-pending signal IDs in the signaling round
- `grogu_conversion_error_signal_ids` (Gauge): Number of signal IDs that failed to convert to signal prices in the signaling round
- `grogu_signal_not_found` (Gauge): number of signal IDs that aren't found from the price list
- `grogu_non_urgent_unavailable_signal_ids` (Gauge): Number of signal IDs that the signal price whose status is unavailable and isn't urgent in the signaling round
- `grogu_filtered_signal_ids` (Gauge): Number of signal IDs that is allowed to submit to the BandChain in the signaling round
- `grogu_signal_price_status` (Gauge): Number of signal prices with specific status
  - Labels: `signal_price_status`

### Submitter

- `grogu_submitting_tx_count` (Counter): Number of times the submitter submits transactions
- `grogu_submit_tx_failed_count` (Counter): Number of times the submitter fail to submit transactions
- `grogu_submit_tx_success_count` (Counter): Number of times the submitter successfully submits transactions
- `grogu_submit_tx_duration` (Summary): Time being consumed for submitting a transaction to the BandChain
  - Percentiles: 50th, 90th, 99th
- `grogu_waiting_sender_duration` (Summary): Time being consumed for waiting the free sender for sending a transaction
  - Percentiles: 50th, 90th, 99th
- `grogu_updated_signal_interval` (Summary): Time interval between the last two updates of the same signal price
  - Labels: `signal_id`
  - Percentiles: 50th, 90th, 99th

## Grafana Dashboard

Grogu provides a pre-built Grafana Dashboard to visualize Grogu metrics efficiently. You can download and import the dashboard from Grafana's official repository.

### Dashboard Link

[Grogu Grafana Dashboard #23217](https://grafana.com/grafana/dashboards/23217-grogu/)

## Metrics Server Configuration

The metrics server can be configured using:

- Configuration file: `metrics-listen-addr` setting
- Command line: `--metrics-listen-addr` flag

The server exposes metrics at the `/metrics` endpoint in Prometheus format.