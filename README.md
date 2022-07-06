# Avalanche Plugin Manager (apm)
Command-line tool to manage virtual machines binaries for [Avalanche](https://github.com/ava-labs/avalanchego). 

## Installation

### From Source
To build from source, you can use the provided build script from the repository root.
```
./scripts/build.sh
```
The resulting `apm` binary will be available in `./build/apm`.

## Commands

### install-vm
Installs a virtual machine by its alias. 

If multiple matches are found (e.g `repository-1/foovm`, `repository-2/foovm`), you will be required to specify the
fully qualified name of the virtual machine to disambiguate the repository to install from.

This will install the virtual machine binary to your `avalanchego` plugin path.

```shell
apm install-vm --vm-alias spacesvm
```

#### Parameters:
- `--vm-alias`: The alias of the VM to install.


### join-subnet
Joins a subnet by its alias.

This will install dependencies for the subnet by calling `install-vm` on each virtual machine required by the subnet.

If multiple matches are found (e.g `repository-1/foo`, `repository-2/foo`), you will be required to specify the
fully qualified name of the subnet definition to disambiguate the repository to install from.


```shell
apm join-subnet --subnet-alias spaces
```

#### Parameters:
- `--subnet-alias`: The alias of the VM to install.

### list-repositories
Lists all tracked repositories.

```shell
apm list-repositories
```

### uninstall-vm
Installs a virtual machine by its alias.

If multiple matches are found (e.g `repository-1/foovm`, `repository-2/foovm`), you will be required to specify the
fully qualified name of the virtual machine to disambiguate the repository to install from.

This will remove the virtual machine binary from your `avalanchego` plugin path.

```shell
apm uninstall-vm --vm-alias spacesvm
```

#### Parameters:
- `--vm-alias`: The alias of the VM to uninstall.

### update

Fetches the latest plugin definitions from all tracked repositories and updates any stale virtual machines.

This will update any virtual machine binaries in your `avalanchego` plugin path with the latest synced definitions.

```shell
apm list-repositories
```


## Examples

### Setting up Credentials for a Private Plugin Repository
You'll need to specify the `--credentials-file` flag which contains your github personal access token. 

Example token file:
```
username: joshua-kim (for GitHub, this field doesn't matter. You can use your username as a placeholder)
password: <personal access token here>
```

Example command to download a subnet's VMs from a private repository:
```
apm join-subnet --subnet-alias=foobar --credentials-file=/home/joshua-kim/token --plugin-path=/home/joshua-kim/dev/avalanchego/build/plugins
```
