# node-manager-plugin

Node manager is a [massa station](https://station.massa.net/) plugin that allows to launch massa node on your machine without having to use terminal. Via it's web interface you can manage staking, get logs choose the network...

> *Note: This plugin need to have **massa wallet plugin** installed*


## setup
install dev dependencies:
```
task install
```

Run go generate and create mocks:
```
task generate
```

Download massa node node binaries:
```
task setup-node-folder
```

Build
```
task build
```

Install the plugin on massa station localy:
```
task install-plugin
```
If massa station was already running, you need to relaunch it for the plugin to be detected.

### standalone
Node manager plugin is expected to be used via massa station.
However, during dev process, you don't necessary need to install the plugin on station each time you want to test it.
To launch the plugin without having to install it in station, you can build it in standalone mode:
```
task build-standalone
```

Then you can run it:
```
task run
```

Then open your browser and go to : `localhost:8080`

> *Even if the plugin run outside of massa station, it still need massa station to be launched on your machine. If massa station is not running, the plugin will be stucked on 'loadling'*


