# node-manager-plugin

Node manager is a [massa station](https://station.massa.net/) plugin that allows you to launch a Massa node on your machine without having to use the terminal. Via its web interface, you can manage staking, get logs, choose the network, etc.

> *Note: This plugin needs to have the **Massa Wallet plugin** installed.*

## Requirements
- [Taskfile](https://taskfile.dev/)
- Node
- go


## Setup
Install dev dependencies:
```
task install
```

Run go generate and create mocks:
```
task generate
```

Download Massa node binaries:
```
task setup-node-folder
```

Build:
```
task build
```

Install the plugin on Massa Station locally:
```
task install-plugin
```
If Massa Station was already running, you need to relaunch it for the plugin to be detected.

### Standalone
Node manager plugin is expected to be used via Massa Station.
However, during the dev process, you don't necessarily need to install the plugin on Station each time you want to test it.
To launch the plugin without having to install it in Station, you can build it in standalone mode:
```
task build-standalone
```

Then you can run it:
```
task run
```

Then open your browser and go to: `localhost:8080`

> *Even if the plugin runs outside of Massa Station, it still needs Massa Station to be launched on your machine. If Massa Station is not running, the plugin will be stuck on 'loading'.*


## Architecture
### Back
The core logic of the node manager is placed in the [core](./int/core/) folder. 
Interactions with other elements (massa-node cli, file system, API handlers, local database...) are handled in specific packages. 
[API handlers](./int/api/handlers/) call core logic which, in turn, calls other drivers.

### Front
The frontend is a TypeScript React project using [vite](https://vite.dev/) framework.
The front end is built and embedded in the Go code, then served by the api (see [html](./int/api/html/html.go) package). 


## Test
Unit tests are created with [testify](https://github.com/stretchr/testify).
Dependencies are mocked with the [mockery](github.com/vektra/mockery) mock generator.
To generate a mock of your interface, set it in [.mockery.yml](.mockery.yml) following the example of other mocked elements.
