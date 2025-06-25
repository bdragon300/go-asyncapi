---
title: "Client application generation"
weight: 340
description: "Building a no-code client application with go-asyncapi"
---

# Client application

One of the `go-asyncapi` core features is building the CLI client application from an AsyncAPI document without writing any code.

The built client application runs on any machine, requires no other libraries preinstalled and has out-of-the-box 
basic pub/sub functionality.
Also, it works well with server variables, channel parameters, bindings, etc. 
And yet, it contains the entities defined in the AsyncAPI document.

To build the client application, you need the [Go toolchain](https://go.dev/doc/install) installed on your machine.

The source code of the client application can also be customized in templates.
See [templating guide]({{<relref "/templating-guide">}}) for more details.

## Usage

{{% details title="Application help output" %}}
```
$ ./client --help                                  
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

## Behavior

Client application opens only one channel/operation to one server at a time. Its default behavior is:

* Accept the data from the (stdin), and outputs the data to the standard output (stdout). Can be changed with the `--file` option.
* Processes only one message at a time, and exits after that. Can be changed with the `--multiple` option.
* If `--multiple` is specified, the delimiter that is used to separate the messages is a newline character. 
  Can be changed with the `--end-of-message` option.
* No timeout is set for the command execution, so it runs indefinitely. You can set a timeout with the `--run-timeout` option.
* Console logging is disabled by default. You can enable the debug logging with the `--debug` option, log output will be sent to stderr.

### Publishing messages

By default, after the client application is started, it waits the input on its stdin.

Here is how to publish a message to the `lightTurnOn` channel using the `turnOn` operation and the `production` 
server with `streetlightId` channel parameter set to `123`:

```
echo "Hello world!" | ./client publish turn-on production --streetlight-id 123
```

### Subscribing to messages

To subscribe to the `lightingMeasured` channel using the `receiveLightMeasurement` operation and 
the `production` server with `streetlightId` channel parameter set to `123`, you can run the following command 
(it will wait for the messages and exit once the first message is received):

```
./client subscribe receive-light-measurement production --streetlight-id 123
```
