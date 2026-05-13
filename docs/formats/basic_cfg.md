# basic.cfg Reference

## Purpose

`basic.cfg` configures Arma 3 server network performance. It controls packet sizing,
bandwidth estimation, and update frequency for both server-to-client and
client-to-server traffic.

ZEUS generates this file at `<install_dir>/.zeus/<profile_name>/configs/basic.cfg`
from the `basic:` section of a profile YAML.

## File Loading

The filename is arbitrary. The path is passed via the `-cfg` startup parameter.
If `-cfg` is not specified, `Arma3.cfg` in the user profile folder is loaded instead.

> **Note:** `basic.cfg` settings apply to clients as well as servers when passed with
> `-cfg` on the client executable.

## Parameter Reference

### Message and Packet Sizing

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MaxMsgSend = 128;` | `128` | Maximum aggregate message packets sent per simulation cycle. Increase on high-upload servers to reduce lag. |
| `MaxSizeGuaranteed = 512;` | `512` | Max payload (bytes) of a guaranteed packet, excluding headers. Guaranteed packets carry non-repetitive events (shots fired, actions). |
| `MaxSizeNonguaranteed = 256;` | `256` | Max payload (bytes) of a non-guaranteed packet. Non-guaranteed packets carry repetitive state (position updates). Increasing may reduce bandwidth but increase lag. |

### Bandwidth Estimation

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MinBandwidth = 131072;` | `131072` | Guaranteed server bandwidth in bps. Overly optimistic values increase lag and CPU load. |
| `MaxBandwidth = 10000000000;` | (unknown) | Upper bandwidth limit in bps. Helps server avoid over-sending. |

### Network Error Thresholds

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MinErrorToSend = 0.001;` | `0.001` | Minimum positional error before sending an update to remote clients. Lower values = smoother distant unit movement at cost of traffic. |
| `MinErrorToSendNear = 0.01;` | `0.01` | Minimum error for nearby units. Independent absolute limit; prevents excessive updates for units moving very slightly. |

**MinErrorToSend formula:**

For a unit at distance `d` with threshold `METS`:
```
update sent when error E >= METS
d = sqrt[(20 × E) / METS]
```

Example: `d = 1000m`, `METS = 0.001` → update when unit moves `50m`.

**MinErrorToSendNear** provides an absolute floor independent of distance.
Without it, very close units would generate excessive updates because the
distance-scaled formula produces very small thresholds at short range.

### Networking Class

```cpp
class sockets
{
    maxPacketSize = 1400;
};
```

| Parameter | Default | Description |
|-----------|---------|-------------|
| `maxPacketSize` | `1400` | Maximum UDP packet size in bytes. Can be set independently per side (client-to-server, server-to-client). Only change if your router/ISP enforces a lower MTU. |

> **Warning:** Setting `MaxSizeGuaranteed` or `MaxSizeNonguaranteed` above `maxPacketSize`
> causes desync.

### Custom Content Limits

| Parameter | Default | Description |
|-----------|---------|-------------|
| `MaxCustomFileSize = 1024;` | (unknown) | Max size (bytes) for player custom face/sound files. Players exceeding this are kicked on connect. `0` = no custom files allowed. |

## Simulation Cycle

Each server simulation cycle (equivalent to a rendering frame):

1. Simulate all unit AI
2. Simulate unit movement, collisions, physics
3. Detect unit visibility
4. Run scripts and FSMs
5. Receive network updates for remote entities
6. Send network updates to other clients

`MaxMsgSend` controls how many packets can be dispatched in step 6.

The server is considered overloaded below 20 cycles/second. Maximum is 50 cycles/second
(configurable via `-limitFPS`, see [Startup Parameters](../protocols/startup_params.md)).

## Tuning Guide

**For a 1024 kbps server:**
```
MaxMsgSend = 256;
MinBandwidth = 768000;
```

**Diagnosing bandwidth usage:**

Use the `#monitor` admin command while connected. If the server has spare bandwidth
headroom, increase `MaxMsgSend` and `MinBandwidth` to utilize it.

**Effect of client-side basic.cfg:**

Client settings affect what the client sends to the server (error computation differs;
camera position is not considered). Both sides can benefit from tuning.

## Example

```
MinBandwidth = 131072;
MaxBandwidth = 10000000000;

MaxMsgSend = 128;
MaxSizeGuaranteed = 512;
MaxSizeNonguaranteed = 256;

MinErrorToSend = 0.001;
MinErrorToSendNear = 0.01;

MaxCustomFileSize = 0;
```

## Related

- [server.cfg Reference](server_cfg.md)
- [Startup Parameters](../protocols/startup_params.md)
