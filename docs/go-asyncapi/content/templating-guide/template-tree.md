---
title: "Template tree"
weight: 930
description: "Tree organization of templates used by go-asyncapi to generate the code, client application, IaC files, etc."
---

# Template tree

All templates, that produce the tool's output, are organized in a tree structure. 

## Code

Legend:

* `*` - template is optional. Executed if found, no error is raised if not.
* `<protocol>` - protocol name, e.g. `amqp`, `kafka`, etc.
* `<implementation>` - implementation name, e.g. `github.com/twmb/franz-go`

Template files referenced by file name:

```
code/
├── main.tmpl
├── operation.tmpl
├── channel.tmpl
├── parameter.tmpl
├── preamble.tmpl
├── server.tmpl
├── message.tmpl
└── schema.tmpl
```

Templates referenced by name (nested templates):

```
code/
├── correlationid.tmpl
│   ├── code/correlationID/incoming/setter
│   ├── code/correlationID/incoming/varType
│   ├── code/correlationID/outgoing/getter
│   └── code/correlationID/outgoing/varType
├── lang/
│   ├── goarray.tmpl
│   │   ├── code/lang/goarray/definition
│   │   └── code/lang/goarray/usage
│   ├── gomap.tmpl
│   │   ├── code/lang/gomap/definition
│   │   └── code/lang/gomap/usage
│   ├── gopointer.tmpl
│   │   ├── code/lang/gopointer/definition
│   │   └── code/lang/gopointer/usage
│   ├── gosimple.tmpl
│   │   ├── code/lang/gosimple/definition
│   │   └── code/lang/gosimple/usage
│   ├── gostruct.tmpl
│   │   ├── code/lang/gostruct/definition
│   │   └── code/lang/gostruct/usage
│   ├── gotypedefinition.tmpl
│   │   ├── code/lang/gotypedefinition/definition
│   │   └── code/lang/gotypedefinition/usage
│   ├── gounion.tmpl
│   │   ├── code/lang/gounion/definition
│   │   └── code/lang/gounion/usage
│   └── govalue.tmpl
│       ├── code/lang/gotypedefinition/value_expr
│       └── code/lang/govalue/usage
└── proto/
    ├── proto_channel.tmpl
    │   ├── code/proto/channel/commonMethods
    │   ├── code/proto/channel/newFunction
    │   ├── code/proto/channel/openFunction
    │   ├── code/proto/channel/publishMethods
    │   ├── code/proto/channel/serverInterface
    │   └── code/proto/channel/subscribeMethods
    ├── proto_message.tmpl
    │   ├── code/proto/message/commonMethods
    │   ├── code/proto/message/decoder
    │   ├── code/proto/message/encoder
    │   ├── code/proto/message/marshalMethods
    │   └── code/proto/message/unmarshalMethods
    ├── proto_mime.tmpl
    │   ├── code/proto/mime/messageDecoder/<mime> *
    │   ├── code/proto/mime/messageDecoder/default
    │   ├── code/proto/mime/messageEncoder/<mime> *
    │   └── code/proto/mime/messageEncoder/default
    ├── proto_operation.tmpl
    │   ├── code/proto/operation/commonMethods
    │   ├── code/proto/operation/newFunction
    │   ├── code/proto/operation/openFunction
    │   ├── code/proto/operation/publishMethods
    │   ├── code/proto/operation/serverInterface
    │   └── code/proto/operation/subscribeMethods
    ├── proto_server.tmpl
    │   ├── code/proto/server/channelOpenMethods
    │   ├── code/proto/server/commonMethods
    │   ├── code/proto/server/connectFunctions
    │   ├── code/proto/server/newFunction
    │   └── code/proto/server/operationOpenMethods
    └── <protocol>/
        ├── code/proto/<protocol>/channel/newFunction/block1 *
        ├── code/proto/<protocol>/channel/publishMethods/block1 *
        ├── code/proto/<protocol>/channel/publishMethods/block2 *
        ├── code/proto/<protocol>/server/impl/<implementation>/connectConsumerFunction *
        ├── code/proto/<protocol>/server/impl/<implementation>/connectFunction *
        ├── code/proto/<protocol>/server/impl/<implementation>/connectProducerFunction *
        ├── code/proto/<protocol>/operation/publishMethods/block1 *
        └── code/proto/<protocol>/operation/publishMethods/block2 *
```

## Client application

Legend:

* `*` - template is optional. Executed if found, no error is raised if not.
* `<protocol>` - protocol name, e.g. `amqp`, `kafka`, etc.
* `<implementation>` - implementation name, e.g. `github.com/twmb/franz-go`

Template files referenced by file name:

```
client/
├── main.tmpl
├── utils.tmpl
└── go.mod.tmpl
```

Templates referenced by name (nested templates):

```
client/
├── channel.tmpl
│   ├── client/channel
│   ├── client/operation
│   └── client/pubsub/proto/serverPubSub
└── proto/
    ├── client/channeloperation/<protocol>/<implementation>/consumer/connect *
    ├── client/channeloperation/<protocol>/<implementation>/producer/connect *
    ├── client/channeloperation/<protocol>/<implementation>/setup *
    ├── client/message/<protocol>/<implementation>/publish *
    ├── client/server/<protocol>/cliMixin *
    ├── client/operation/<protocol>/publish/prepareEnvelope *
    └── client/channel/<protocol>/publish/prepareEnvelope *
```

## Infrastructure files

Legend:

* `*` - template is optional. Executed if found, no error is raised if not.
* `<protocol>` - protocol name, e.g. `amqp`, `kafka`, etc.
* `<engine>` - engine name, e.g. `docker`, etc.
* `<section>` - section name in resulted file. For `docker` engine it's `services`, `volumes`, etc.

Template files referenced by file name:

```
infra/
└── main.tmpl
```

Templates referenced by name (nested templates):

```
infra/
└── <engine>/
    ├── infra/<engine>/<protocol>/<section> *
    ├── infra/<engine>/<protocol>/<section>/extra *
    └── infra/<engine>/<protocol>/extra *
```
