# Avalanche Plugin Manager (apm)

Note: This code is currently in Alpha. Proceed at your own risk.

Avalanche Plugin Manger is a command-line tool to manage virtual machines binaries for
[avalanchego](https://github.com/ava-labs/avalanchego).

`apm` allows users to build their own custom repositories to provide virtual machine and subnet definitions outside of
the [avalanche-plugins-core](https://github.com/ava-labs/avalanche-plugins-core) repository. Core ships with the `apm`,
but users have the option of adding their own using the `add-repository` command.

## Installation

### Source
If you are planning on building from source, you will need [golang](https://go.dev/doc/install) >= 1.18.x installed.

To build from source, you can use the provided build script from the repository root.
```
./scripts/build.sh
```
The resulting `apm` binary will be available in `./build/apm`.

## Commands

### add-repository
Starts tracking a plugin repository.

```shell
apm add-repository --alias foo --url https://github.com/joshua-kim/foobar
```

#### Parameters:
- `--alias`: The alias of the VM to install.
- `--url`: The alias of the VM to install.
 
### install-vm
Installs a virtual machine by its alias. 

If multiple matches are found (e.g `repository-1/foovm`, `repository-2/foovm`), you will be required to specify the
fully qualified name of the virtual machine to disambiguate the repository to install from.

This will install the virtual machine binary to your `avalanchego` plugin path.

```shell
apm install-vm --vm spacesvm
```

#### Parameters:
- `--vm`: The alias of the VM to install.


### join-subnet
Joins a subnet by its alias.

This will install dependencies for the subnet by calling `install-vm` on each virtual machine required by the subnet.

If multiple matches are found (e.g `repository-1/foo`, `repository-2/foo`), you will be required to specify the
fully qualified name of the subnet definition to disambiguate the repository to install from.


```shell
apm join-subnet --subnet spaces
```

#### Parameters:
- `--subnet`: The alias of the VM to install.

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
apm uninstall-vm --vm spacesvm
```

#### Parameters:
- `--vm`: The alias of the VM to uninstall.

### update

Fetches the latest plugin definitions from all tracked repositories.


```shell
apm list-repositories
```

### Upgrade 

Upgrades a virtual machine binary. If one is not provided, this will upgrade all virtual machine binaries in your
`avalanchego` plugin path with the latest synced definitions.

For a virtual machine to be upgraded, it must have been installed using the `apm`.

```shell
apm upgrade
```

#### Parameters
- `--vm`: (Optional) The alias of the VM to upgrade. If none is provided, all VMs are upgraded.

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
apm join-subnet --subnet=foobar --credentials-file=/home/joshua-kim/token
```
