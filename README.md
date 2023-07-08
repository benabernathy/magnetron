# Magnetron
_____
Magnetron is a modern, cross-platform implementation of the Hotline tracker
protocol. It is written in Golang.

The goal of Magnetron is to provide flexibility for those who wish to provide
trackers for Hotline servers. Magnetron can run on any platform that Golang supports.
Examples include Windows, Linux (x86 & arm), Raspberry Pi, and MacOs (x86 and arm). 

Without Hotline trackers, it is very difficult to locate Hotline servers. The original
tracker was created to allow those with dynamic IPs to easily advertise their local Hotline
server. Magnetron aims to keep this aspect of the community alive by helping "future proof"
a critical, and often overlooked aspect of the Hotline community.

There are alternative trackers out there, such as the excellent Pitbull tracker, however
implementing a tracker in Golang provides much more flexibility. Flexibility will help maintain
and perhaps even grow the community.

## Getting Started
____

Magnetron is currently in alpha release and as such, we do not provide prebuilt binaries. Magnetron
will provide prebuilt binaries and Docker images once it has reached the beta stage.

### Building
A Makefile is provided for your convenience, although Make is not required.
You can build Magnetron by running the following command.
```shell
make all
```

Alternatively you build the application by running the following command.
```shell
go build cmd/magnetron/main.go -o magnetron 
```

### Configuration
In order to run, Magnetron requires a configuration. The default configuration
provided by the application should be considered the minimal set of configuration
required to serve as a tracker. To generate a configuration file with default values,
run the following command.

```shell
magnetron config init config.yml
```

The previous command will tell Magnetron to write out the default configuration to a
file named `config.yml` in your current directory. As expected, providing a different
file name or path will cause the file to be written with the specified name or location.

### Running Magnetron
___

Magnetron can be run by telling it to run the server and use a specified configuration
file. For example, you can use the following command to run the server.

```shell
magnetron serve config.yml
```

Magnetron will then start up and listen for server and client connections. 

### Docker
___
Docker is not official supported yet, but a basic Dockerfile can be found in this directory. It can also be built
using the makefile command: 

```shell
make docker-build
```

Prebuilt Docker images will be available shortly.
