# apm
Command-line tool to manage virtual machines and subnets for Avalanche. 

## Installation

### From Source
To build from source, you can use the provide build script.
```
./scripts/build.sh
```
The resulting `apm` binary will be available in `./build/apm`.

## Examples
Currently, this repo is in alpha. Since the core repository is a private repo, you'll need to specify the `--credentials-file` flag which contains your github personal access token. 

Example token file:
```
username: joshua-kim (for GitHub, this field doesn't matter. You can use your username as a placeholder)
password: <personal access token here>
```

Download the vms for a subnet and whitelisted it locally (make sure your admin api server is running on `127.0.0.1:9650`!)
```
./build/apm join-subnet --subnet-alias=spaces --credentials-file=/Users/joshua.kim/token --plugin-path=/Users/joshua.kim/github/avalanchego-internal/build/plugins
```
