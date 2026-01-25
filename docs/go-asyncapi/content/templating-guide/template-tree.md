---
title: "Template tree"
weight: 930
description: "Tree organization of templates used by go-asyncapi to generate the code, client application, IaC files, etc."
---

# Template tree

All templates that are used in `go-asyncapi` are organized in a tree structure, giving the user a way to customize 
the final generation result on any granularity level.

## Code

Legend:

* `*` - template is optional, no error is raised if not found.
* `<protocol>` - protocol name, e.g. `amqp`, `kafka`, etc.
* `<implementation>` - implementation name, e.g. `github.com/twmb/franz-go`
* `<jsonschema_type>` - JSON Schema type, e.g. `string`, `integer`, etc.
* `<jsonschema_format>` - JSON Schema format, e.g. `date-time`, `uuid`, etc.
* `<security_type>` - type of security scheme, e.g. `apiKey`, `userPassword`, etc.
* `<mime>` - MIME type, e.g. `application/json`, `application/xml`, etc.

The following templates generate the code:

```
main.tmpl
operation.tmpl
channel.tmpl
parameter.tmpl
preamble.tmpl
server.tmpl
message.tmpl
proto_channel.tmpl
proto_message.tmpl
proto_operation.tmpl
proto_server.tmpl
schema.tmpl
security.tmpl
<protocol>_channel.tmpl
<protocol>_message.tmpl
<protocol>_operation.tmpl
<protocol>_server.tmpl
unknown_channel.tmpl
unknown_message.tmpl
unknown_operation.tmpl
unknown_server.tmpl
code/
├── runtimeExpression/
│   ├── code/runtimeExpression/setterBody
│   └── code/runtimeExpression/getterBody
├── security/
│   └── code/security/<security_type> *
├── lang/
│   ├── goarray/
│   │   ├── code/lang/goarray/definition
│   │   └── code/lang/goarray/usage
│   ├── gomap/
│   │   ├── code/lang/gomap/definition
│   │   └── code/lang/gomap/usage
│   ├── gopointer/
│   │   ├── code/lang/gopointer/definition
│   │   └── code/lang/gopointer/usage
│   ├── gosimple/
│   │   ├── code/lang/gosimple/definition
│   │   └── code/lang/gosimple/usage
│   ├── gostruct/
│   │   ├── code/lang/gostruct/definition
│   │   └── code/lang/gostruct/usage
│   ├── gotypedefinition/
│   │   ├── code/lang/gotypedefinition/definition
│   │   ├── code/lang/gotypedefinition/usage
│   │   └── code/lang/gotypedefinition/value_expr
│   ├── gounion/
│   │   ├── code/lang/gounion/definition
│   │   └── code/lang/gounion/usage
│   └── typeFormat/<jsonschema_type>/<jsonschema_format>/
│       ├── code/lang/typeFormat/<jsonschema_type>/<jsonschema_format>/definition *
│       └── code/lang/typeFormat/<jsonschema_type>/<jsonschema_format>/usage *
└── proto/
    ├── channel/
    │   ├── code/proto/channel/commonMethods
    │   ├── code/proto/channel/newFunction
    │   ├── code/proto/channel/openFunction
    │   ├── code/proto/channel/publishMethods
    │   ├── code/proto/channel/serverInterface
    │   └── code/proto/channel/subscribeMethods
    ├── message/
    │   ├── code/proto/message/commonMethods
    │   ├── code/proto/message/decoder
    │   ├── code/proto/message/encoder
    │   ├── code/proto/message/marshalMethods
    │   └── code/proto/message/unmarshalMethods
    ├── mime/
    │   ├── code/proto/mime/messageDecoder/<mime> *
    │   ├── code/proto/mime/messageDecoder/default
    │   ├── code/proto/mime/messageEncoder/<mime> *
    │   └── code/proto/mime/messageEncoder/default
    ├── operation/
    │   ├── code/proto/operation/commonMethods
    │   ├── code/proto/operation/publishMethods
    │   ├── code/proto/operation/openFunction
    │   ├── code/proto/operation/serverInterface
    │   ├── code/proto/operation/securityInterface
    │   ├── code/proto/operation/subscribeMethods
    │   └── operationReply/
    │       ├── code/proto/operation/operationReply/commonMethods
    │       ├── code/proto/operation/operationReply/publishMethods
    │       └── code/proto/operation/operationReply/subscribeMethods
    ├── server/
    │   ├── code/proto/server/channelOpenMethods
    │   ├── code/proto/server/commonMethods
    │   ├── code/proto/server/connectFunctions
    │   ├── code/proto/server/newFunction
    │   ├── code/proto/server/operationOpenMethods
    │   └── code/proto/server/securityInterface
    └── <protocol>/
        ├── channel/
        │   ├── code/proto/<protocol>/channel/bindings/values *
        │   ├── code/proto/<protocol>/channel/newFunction/block1 *
        │   ├── code/proto/<protocol>/channel/publishMethods/block1 *
        │   └── code/proto/<protocol>/channel/publishMethods/block2 *
        ├── message/
        │   └── code/proto/<protocol>/message/bindings/values *
        ├── operation/
        │   └── code/proto/<protocol>/operation/bindings/values *
        └── server/
            ├── code/proto/<protocol>/server/bindings/values *
            └── impl/<implementation>/
                ├── code/proto/<protocol>/server/impl/<implementation>/connectConsumerFunction *
                ├── code/proto/<protocol>/server/impl/<implementation>/connectFunction *
                └── code/proto/<protocol>/server/impl/<implementation>/connectProducerFunction *
```

## Client application

Legend:

* `*` - template is optional. Executed if found, no error is raised if not.
* `<protocol>` - protocol name, e.g. `amqp`, `kafka`, etc.
* `<implementation>` - implementation name, e.g. `github.com/twmb/franz-go`
* `<security_type>` - type of security scheme, e.g. `apiKey`, `userPassword`, etc.

The following templates generate the client application using the code generated by code templates:

```
main.tmpl
utils.tmpl
go.mod.tmpl
client/
├── channel/
│   ├── client/channel
│   └── client/channel/<protocol>/publish/prepareEnvelope *
├── operation/
│   ├── client/operation
│   └── client/operation/<protocol>/publish/prepareEnvelope *
├── channeloperation/
│   ├── client/channeloperation/<protocol>/<implementation>/consumer/connect *
│   ├── client/channeloperation/<protocol>/<implementation>/producer/connect *
│   └── client/channeloperation/<protocol>/<implementation>/setup *
├── security/
│   ├── client/security/<security_type>/cmdFlags *
│   ├── client/security/<security_type>/server/getCredentials *
│   └── client/security/<security_type>/operation/getCredentials *
├── pubsub/
│   ├── client/pubsub/proto/serverPubSub
│   └── client/pubsub/proto/serverPubSub/open
├── client/operationReply
├── client/message/<protocol>/<implementation>/publish *
└── client/server/<protocol>/cliMixin *
```

## Infrastructure files

Legend:

* `*` - template is optional. Executed if found, no error is raised if not.
* `<protocol>` - protocol name, e.g. `amqp`, `kafka`, etc.
* `<engine>` - engine name, e.g. `docker`, etc.
* `<section>` - section name in resulted file. For `docker` engine it's `services`, `volumes`, etc.

The following templates generate the infrastructure files:

```
main.tmpl
infra/
└── <engine>/<protocol>/
    ├── infra/<engine>/<protocol>/<section> *
    ├── infra/<engine>/<protocol>/<section>/extra *
    └── infra/<engine>/<protocol>/extra *
```

## Diagrams

The following templates generate the D2 diagram code:

```
main.tmpl
diagram/
├── diagram/document
├── diagram/channel
├── diagram/operation
└── diagram/server
```

# UI

The following templates generate the HTML file:

```
main.tmpl
```
