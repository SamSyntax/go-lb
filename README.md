# Go Load Balancer with Configurable Algorithms and External Server Configuration

This Go project implements a load balancer that supports two balancing methods: Round Robin (`rr`) and Weighted Round Robin (`wrr`). It allows you to spawn local servers or read external server addresses from a configuration file (`.json` or `.yaml`) via a flag.

## Features

- **Round Robin** (`rr`) and **Weighted Round Robin** (`wrr`) algorithms for load balancing.
- Support for **local server spawning** or **external server configuration** via `.json` or `.yaml` files.
- Configurable via command-line flags.

## Command-Line Flags

- `-amount`: Number of local servers to spawn (used only with `-env local`). Default is `1234`.
- `-method`: Load balancing method. Choose between:
  - `rr`: Round Robin.
  - `wrr`: Weighted Round Robin.
- `-env`: Environment setting. Choose between:
  - `local`: Spawns the specified amount of local servers.
  - `external`: Reads server addresses from an external file (provided via `-path` flag).
- `-path`: Specifies the path to the external `.json` or `.yaml` configuration file (used with `-env external`).

### Example Usage

1.**Spawn Local Servers** (5 servers, round-robin method):
```bash
go run *.go -amount 5 -method rr -env local
```
Use External Servers from a JSON File (weighted round-robin method):

```bash
go run *.go -method wrr -env external -path ./servers.json
```
Use External Servers from a YAML File (round-robin method):

```bash
go run *.go -method rr -env external -path ./servers.yaml
```
### Configuration File Format
When using the -env external flag, the load balancer will read server information from a configuration file. You can provide the file in either YAML or JSON format.

## Sample YAML Configuration (servers.yaml)
```yaml
---
- addr: https://facebook.com
  weight: 2
- addr: https://twitch.tv
  weight: 1
- addr: https://google.com
  weight: 3
```
## Sample JSON Configuration (servers.json)
```json
[
  {
    "addr": "https://facebook.com",
    "weight": 2
  },
  {
    "addr": "https://twitch.tv",
    "weight": 1
  },
  {
    "addr": "https://google.com",
    "weight": 3
  }
]
```
## How It Works
### Load Balancing Methods
 Round Robin (rr): Distributes requests evenly across all available servers.
Weighted Round Robin (wrr): Distributes requests based on the weight assigned to each server. Servers with higher weights receive more traffic.
Local Server Spawning
When using -env local, the program spawns a number of local servers on ports starting from 8000 (e.g., localhost:8000, localhost:8001, etc.).

# External Servers
When using -env external with the -path flag, the load balancer reads external server addresses from the specified JSON or YAML file and balances requests accordingly.

## Example Output
Example output when running with 3 local servers:

```bash
serving requests at localhost:7000
forwarding to "localhost:8000"
forwarding to "localhost:8001"
forwarding to "localhost:8002"
```
Example output when using an external JSON configuration:

```bash
serving requests at localhost:7000
forwarding to "https://facebook.com"
forwarding to "https://twitch.tv"
forwarding to "https://google.com"
```

## Port flag
We can pass <b>-srv-port</b> flag to specify port for local servers and <b>-port</b> flag to specify port for load balancer
```bash
./lb -srv-port 8000 -port 7000 
```

## Health check
We are able to run health check on external servers listed in the config file
```bash
./lb -healthCheck -path ./servers.yaml
```

# License
This project is open-source and available under the MIT License.

