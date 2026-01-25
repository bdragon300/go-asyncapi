---
title: "client"
weight: 320
description: "Generating the client application executable"
---

# Building the client application

{{% hint important %}}
`client` command requires the [Go toolchain](https://go.dev/doc/install) installed on your machine.
{{% /hint %}}

`client` command builds a zero-code CLI application binary based only on an AsyncAPI document. This application represents the 
AsyncAPI entities as its commands and options, allowing to publish and subscribe messages to/from the specified channels.
It also supports all features that `go-asyncapi` supports: server variables, channel parameters, bindings, etc.

Under the hood, `client` command generates the boilerplate code using the [code]({{< relref "/commands/code" >}}) command,
then it generates the application code that uses the boilerplate code. On final step, it builds the binary by executing 
the `go build` command from the Go toolchain.

The result is a standalone executable file named `client` (or `client.exe` on Windows) in the current working directory.

{{% hint tip %}}
`go-asyncapi` also passes
the environment variables to `go build` command running inside, so, for example, setting the `GOARCH=arm64` variable
will produce a binary for ARM64 architecture. [More about Go variables](https://go.dev/doc/install/source#environment).
{{% /hint %}}

{{% hint tip %}}
As for the [code]({{< relref "/commands/code" >}}) command, the source code of the client application can
also be customized in templates. See [templating guide]({{<relref "/templating-guide/overview">}}) for more details.
{{% /hint %}}


{{% hint tip %}}
By default, the temporary generated sources used to build the application are removed after the build is finished.
To keep them from removing, use the `--keep-source` option.

The `code` command also has the related option, `--client-app`, that generates the client application source code
next to the boilerplate code in target directory without building the application.
{{% /hint %}}

## Usage

{{% hint tip %}}
All command options are duplicated in [configuration]({{<relref "/configuration">}}) file. So, you can set them there as well.
{{% /hint %}}

To build the client application, run the following command:

```bash
go-asyncapi client <asyncapi-document> [options...]
```

After the command is finished, the built application file will be available in the current working directory.
The general approach to use the built client application is:

```bash
./client <mode> <channel-or-operation> <server> [<channel-parameters>...] [<server-variables>...] [options...]
```

{{% details title="Application help output" %}}
```bash
./client --help                                  
Usage: client [--docker] [--proxy-host PROXY-HOST] [--debug] [--file FILE] [--multiple] [--headers HEADERS] [--end-of-message END-OF-MESSAGE] [--run-timeout RUN-TIMEOUT] <command> [<args>]

Options:
  --docker               Proxy connections to a docker-proxy keeping the original destination port numbers. Proxy host can be specified with --proxy-host
  --proxy-host PROXY-HOST
                         If proxying is enabled, redirect all connections to this host [default: 127.0.0.1]
  --debug, -d            Enable debug logging
  --file FILE, -f FILE   File to read or write message data; - means stdin/stdout [default: -]
  --multiple, -m         Do not exit after the first message processed
  --headers HEADERS      Message header to send; format: key=value [key=value ...]
  --end-of-message END-OF-MESSAGE
                         Delimiter that separates the message payloads in stream. Empty string means EOF (or Ctrl-D in interactive terminal) [default: 
]
  --run-timeout RUN-TIMEOUT
                         Timeout to run the command. By default, the command runs indefinitely
  --help, -h             display this help and exit

Commands:
  subscribe              Subscribe to a channel
  publish                Publish to a channel
```
{{% /details %}}

### Publishing messages

By default, after the client application is started, it awaits the message to publish from stdin. Once a message is read, it is
published to the specified channel and the application exits.

{{% hint default %}}
Here is how to publish a text message to the `lightTurnOn` channel using the `turnOn` operation and the `production`
server with `streetlightId` channel parameter set to `123`:

```bash
echo "Hello world!" | ./client publish turn-on production --streetlight-id 123
```
{{% /hint %}}

### Subscribing to messages

By default, after the client application is started as a subscriber, it waits for messages appearing in the specified
channel. Once a message is received, it is printed to the stdout and the application exits.

{{% hint default %}}
To subscribe to the `lightingMeasured` channel using the `receiveLightMeasurement` operation and
the `production` server with `streetlightId` channel parameter set to `123`:

```bash
./client subscribe receive-light-measurement production --streetlight-id 123
```
{{% /hint %}}

### Behavior

Client application opens only one channel/operation to one server at a time. Its default behavior is following:

* Reads the data from stdin, outputs the data to the stdout. Can be changed with the `--file` option.
* Exits after the first processed message, unless the `--multiple` flag is set. In multiple mode, messages are 
  separated by newline character (customizable by `--end-of-message` option).
* No run timeout. To set a timeout use the `--run-timeout` option.
* Debug logging is disabled by default. Use `--debug` option to enable, log output will be sent to stderr.

### Proxying connections to Docker

The client application can also proxy connections to Docker private network running on the machine. 
To enable proxying, use the `--docker` option.
