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

Magnetron provides prebuilt binaries and Docker images. YOu can find more information about how
to use the Docker images later on in the readme. 

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
In order to run, Magnetron requires a configuration file. The default configuration
provided by the application should be considered the minimal set of configuration
required to serve as a tracker. To generate a configuration file with default values,
run the following command.

```shell
magnetron config init config.yml
```

The previous command will tell Magnetron to write out the default configuration to a
file named `config.yml` in your current directory. As expected, providing a different
file name or path will cause the file to be written with the specified name or location.

#### Password Configuration
If you wish to require servers to use a password to register with your tracker, you'll 
then need to perform the following steps. Magnetron stores hashes of passwords so the 
plain text password is never stored locally. When the server connects and provides its 
password, Magnetron will hash the server-supplied password and check it against the
stored hash. If they match, then the server is registered. Otherwise, the server's 
information will not be registered with the tracker.

Like the configuration, Magnetron provides the ability to generate a default password
configuration file for you to start with. This is done by running the following command.

```shell
magnetron password init passwords.yml
```

The previous command will tell Magnetron to write out the default password configuration to a
file named `passwords.yml` in your current directory. As expected, providing a different
file name or path will cause the file to be written with the specified name or location.

Then you'll need to update your Magnetron configuration file and make sure the following
fields are correct:

```yaml
EnablePasswords: true
PasswordFile: "./passwords.yml"
```
Next you'll need to open the `passwords.yml` file and set entries:

```yaml
PasswordEntries:
  - Name: "default1"
    Description: "default1's password is password"
    Password: "$2a$10$UiFV2qCHvXWeYbhk2LlqueKvQwPqJWTxuJAqUhuCLdz2F9fJr8dNG"
```

The `Name` and `Description` fields can be any value you want. They are there to help you
remember/selectively revoke passwords. The `Password` field is the string value of the 
hashed password. More details on how to generate this are provided later.

You may add as many entries as you like, for example:

```yaml
PasswordEntries:
  - Name: "default1"
    Description: "default1's password is password"
    Password: "$2a$10$UiFV2qCHvXWeYbhk2LlqueKvQwPqJWTxuJAqUhuCLdz2F9fJr8dNG"
  - Name: "default2"
    Description: "default2's password is password"
    Password: "$2a$10$UiFV2qCHvXWeYbhk2LlqueKvQwPqJWTxuJAqUhuCLdz2F9fJr8dNG"
```

The `Name` fields do not have to be unique, although their utility is questionable if there are duplicates...

Magnetron provides a convenience function for encrypting passwords. You can encrypt a password by running:

```shell
magnetron password encrypt
```

Following the on-screen prompts will result in a hashed password. Magnetron can also check a provided
password against the password configuration file. You can do this by running the following command:

```shell
magnetron password check passwords.yml
```

After you follow the on-screen prompts, Magentron will check the supplied password against the password
list and tell you if it is valid or not. This can be a handy debug tool. Magnetron can also perform a 
limited password configuration validation against a supplied password configuration file:

```shell
magnetron validate passwords.yml
```

### Running Magnetron
___

Magnetron can be run by telling it to run the tracker and use a specified configuration
file. For example, you can use the following command to run the tracker.

```shell
magnetron serve config.yml
```

Magnetron will then start up and listen for hotline server and client connections. 

### Building Docker Images
___
A basic Dockerfile can be found in this directory. It can also be built using the make command: 

```shell
make docker-build
```

### Prebuilt Docker Images
Prebuilt Docker images are available for linux (amd64 and arm64). They can be pulled by running the following command

```shell
docker pull benabernathy/magnetron:latest
```

Note: If you are using Docker Desktop (e.g. on Windows), you'll need to replace `$(pwd)` with the fully qualified configuration directory path.

You can externalize Magnetron's configuration by mounting your host filesystem via a volume to the Docker container. To generate a default configuration, use the following command (assuming your desired host file system path is `conf`):

```shell
docker run --rm -v $(PWD)/conf:/usr/local/var/magnetron benabernathy/magnetron:latest "config" "init" "/usr/local/var/magnetron/config.yml"
```

After you have run this command, the config file will now be in your `conf` directory. You'll most likely need to change the `ClientHost` and `ServerHost` values to:

```yaml
ClientHost: 0.0.0.0:5498
ServerHost: 0.0.0.0:5498
```

After making changes, you can re-run Magnetron with your new configuration, publishing the ports to your host:

```shell
docker run --name magnetron --rm -v $(PWD)/conf:/usr/local/var/magnetron -p 5499:5499/udp -p 5498:5498 benabernathy/magnetron:latest
```

If you want to initialize the password configuration, you can run the follwoing command:

```shell
docker run --rm -v $(PWD)/conf:/usr/local/var/magnetron benabernathy/magnetron:latest "password" "init" "/usr/local/var/magnetron/passwords.yml"
```

Note: The path to the passwords.yml file specified in the field `PasswordFile` in this case would be: `/usr/local/var/magnetron/passwords.yml`

