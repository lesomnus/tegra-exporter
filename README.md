# tegra-exporter

Reads [`tegrastats`](https://docs.nvidia.com/jetson/archives/r38.4/DeveloperGuide/AT/JetsonLinuxDevelopmentTools/TegrastatsUtility.html) output and exports metrics via [OpenTelemetry](https://opentelemetry.io/) or via any exporter supported by [mkot](https://github.com/lesomnus/mkot) like [prometheus](https://github.com/lesomnus/mkot/tree/e588c3260503fd56f938fe5b1867c770e3570b18/prometheus).

## Installation

### Download binary from GitHub Releases

```sh
VERSION=260510 \
curl -sSLO https://github.com/lesomnus/tegra-exporter/releases/download/${VERSION}/tegra-exporter \
&& chmod +x tegra-exporter \
&& ./tegra-exporter version
```

Pre-built binaries for `linux/arm64` are attached to each [GitHub Release](https://github.com/lesomnus/tegra-exporter/releases).

### Docker

```sh
docker run \
  -v /sys:/sys:ro \
  -v /sys/kernel/debug:/sys/kernel/debug:ro \
  -v /dev:/dev \
  -v /lib:/lib:ro \
  -v /usr/bin/tegrastats:/usr/bin/tegrastats:ro \
  -v $(pwd)/tegra-exporter.yaml:/tegra-exporter.yaml:ro \
  -it --rm ghcr.io/lesomnus/tegra-exporter:260510
```

### Build from source

```sh
go install github.com/lesomnus/tegra-exporter@latest
```

Requires Go 1.26+.

## Configuration

```yaml
stat: ["tegrastats"]

otel:
  exporters:
    pretty: {}
    prometheus/local:
      endpoint: ":8888"
      namespace: tegra_exporter

  providers:
    meter:
      exporters:
        - prometheus/local
    logger:
      exporters:
        - pretty
```

`tegra-exporter` looks for a config file named `tegra-exporter.yaml` (or `.yml`) in the current working directory.
If no config file is found, it uses default config shown above.

```yaml
otel:
  exporters:
    otlp/local:
      endpoint: localhost:4317
      tls:
        insecure: true

  providers:
    meter:
      processors:
        - resource/tegra-exporter
```

`otlp` exporter can be used to push metrics to a remote OpenTelemetry Collector.


## Metrics

All metrics are exposed as OpenTelemetry gauges.

### Memory

| Metric                | Unit  | Description                    |
| --------------------- | ----- | ------------------------------ |
| `tegra.ram.in_use`    | MB    | RAM in use                     |
| `tegra.ram.total`     | MB    | Total available RAM            |
| `tegra.ram.lfb_count` | count | Number of Largest Free Blocks  |
| `tegra.ram.lfb_size`  | MB    | Size of the Largest Free Block |
| `tegra.iram.in_use`   | MB    | IRAM in use (on-chip SRAM)     |
| `tegra.iram.total`    | MB    | Total available IRAM           |
| `tegra.swap.in_use`   | MB    | Swap in use                    |
| `tegra.swap.total`    | MB    | Total swap                     |
| `tegra.swap.cached`   | MB    | Cached swap                    |

### CPU

| Metric                  | Unit | Attributes | Description                                        |
| ----------------------- | ---- | ---------- | -------------------------------------------------- |
| `tegra.cpu.utilization` | %    | `index`    | Per-core utilization relative to current frequency |
| `tegra.cpu.frequency`   | MHz  | `index`    | Per-core frequency                                 |

The `index` attribute is a zero-based integer string (e.g. `"0"`, `"1"`, ...) identifying the CPU core.

### GPU (GR3D)

| Metric                   | Unit | Attributes | Description                           |
| ------------------------ | ---- | ---------- | ------------------------------------- |
| `tegra.gr3d.utilization` | %    | —          | GPU activation ratio                  |
| `tegra.gr3d.frequency`   | MHz  | `index`    | Per-GPC frequency (`"0"`, `"1"`, ...) |

Jetson AGX Orin has two GPCs; Jetson Thor has three. Each GPC has its own frequency.

### External Memory Controller (EMC)

| Metric                  | Unit | Description                  |
| ----------------------- | ---- | ---------------------------- |
| `tegra.emc.utilization` | %    | Memory bandwidth utilization |
| `tegra.emc.frequency`   | MHz  | EMC frequency                |

### Hardware Engines

Engines with multiple instances carry an `index` attribute (`"0"`, `"1"`, ...).

| Metric                    | Unit | Attributes | Description                           |
| ------------------------- | ---- | ---------- | ------------------------------------- |
| `tegra.vic.utilization`   | %    | —          | Video Image Compositor utilization    |
| `tegra.vic.frequency`     | MHz  | —          | VIC frequency                         |
| `tegra.ape.frequency`     | MHz  | —          | Audio Processing Engine frequency     |
| `tegra.nvenc.utilization` | %    | `index`    | Hardware video encoder utilization    |
| `tegra.nvenc.frequency`   | MHz  | `index`    | NVENC frequency                       |
| `tegra.nvdec.utilization` | %    | `index`    | Hardware video decoder utilization    |
| `tegra.nvdec.frequency`   | MHz  | `index`    | NVDEC frequency                       |
| `tegra.nvdla.utilization` | %    | `index`    | Deep Learning Accelerator utilization |
| `tegra.nvdla.frequency`   | MHz  | `index`    | NVDLA frequency                       |
| `tegra.nvjpg.utilization` | %    | `index`    | JPEG encoder/decoder utilization      |
| `tegra.nvjpg.frequency`   | MHz  | `index`    | NVJPG frequency                       |
| `tegra.ofa.utilization`   | %    | —          | Optical Flow Accelerator utilization  |
| `tegra.ofa.frequency`     | MHz  | —          | OFA frequency                         |

### Temperature & Power

| Metric                | Unit | Attributes | Description                                                              |
| --------------------- | ---- | ---------- | ------------------------------------------------------------------------ |
| `tegra.temperature`   | C    | `sensor`   | Temperature per thermal zone (e.g. `"cpu"`, `"soc2"`)                    |
| `tegra.power.current` | mW   | `sensor`   | Current power consumption per rail (e.g. `"VDD_IN"`, `"VDD_CPU_GPU_CV"`) |
| `tegra.power.average` | mW   | `sensor`   | Average power consumption per rail                                       |

