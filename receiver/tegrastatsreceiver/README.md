# tegrastats receiver

An [OpenTelemetry Collector](https://github.com/open-telemetry/opentelemetry-collector) receiver that runs `tegrastats` inside the Collector,
so no separate `tegra-exporter` process is needed.
It produces the same metrics as `tegra-exporter`; see [Metrics](../../README.md#metrics).

| Status    |         |
| --------- | ------- |
| Stability | alpha   |
| Signals   | metrics |

## Build

Add it to an [ocb](https://github.com/open-telemetry/opentelemetry-collector/tree/main/cmd/builder) manifest:

```yaml
receivers:
  - gomod: github.com/lesomnus/tegra-exporter/receiver/tegrastatsreceiver v0.0.0-<date>-<commit>
```

Use a pseudo-version of a commit on `main`, e.g. from `go list -m github.com/lesomnus/tegra-exporter/receiver/tegrastatsreceiver@main`.

## Configuration

```yaml
receivers:
  tegrastats:
    mode: push
    command: ["tegrastats", "--interval", "5000"]
```

| Field                 | Default          | Description                                                                          |
| --------------------- | ---------------- | ------------------------------------------------------------------------------------ |
| `mode`                | `push`           | `push` or `scrape`; see below.                                                       |
| `command`             | `["tegrastats"]` | Command and arguments to run. `["$fake"]` generates fake stats with a random walk.   |
| `collection_interval` | `1m`             | `scrape` only. How often to emit.                                                    |
| `initial_delay`       | `1s`             | `scrape` only. Delay before the first emission.                                      |
| `timeout`             | `0s`             | `scrape` only. Timeout of each emission; `0s` means no timeout.                      |
| `stale_timeout`       | `10s`            | `scrape` only. A line older than this is not emitted, so a stalled command goes quiet. |

### Modes

- `push` emits every line `tegrastats` prints, so its `--interval` (default `1000` ms) sets the rate.
- `scrape` keeps the latest line and emits it every `collection_interval`, the same way `hostmetrics` does.
  Use it to align with other scrapers, or to sample less often than `tegrastats` prints.

```yaml
receivers:
  tegrastats:
    mode: scrape
    collection_interval: 15s
```

If the command exits, it is restarted after 3 seconds and the reason is logged.
Data points are timestamped when the line is read, because `tegrastats` prints local time without a zone.

## Running in a container

`tegrastats` comes from the host's JetPack, so the Collector container needs what the `tegra-exporter` image needs:

```sh
-v /sys:/sys:ro \
-v /sys/kernel/debug:/sys/kernel/debug:ro \
-v /dev:/dev \
-v /lib:/lib:ro \
-v /usr/bin/tegrastats:/usr/bin/tegrastats:ro
```

Some values, such as GR3D and EMC, are only readable as root.
