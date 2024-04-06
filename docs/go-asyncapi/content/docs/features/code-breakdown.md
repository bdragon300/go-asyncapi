---
title: "Code breakdown"
weight: 330
description: "Generated code could be broken down by packages and files by type, by entity, all-in-one package"
---

# Breakdown the generated code

`go-asyncapi` puts all the generated code to the target directory, breaking it down by the packages and files.
Also, additional packages will be created and placed alongside: `impl` for the protocol implementations and 
`encoding` for the encoders/decoders.

The way how the generated code will breaked down by packages and files is controlled by `--*-scope` cli args. 

By default, the tool creates a package per entity type (**models**, **servers**, **channels**, etc.). 
Every single entity is put to a separate file.

{{< details "Example" >}}
```
target_dir
├── channels
│   ├── channel1.go
│   ├── channel2.go
│   └── ...
├── servers
│   ├── server1.go
│   ├── server2.go
│   └── ...
├── models
│   ├── model1.go
│   ├── model2.go
│   └── ...
├── ...
├── impl
│   ├── kafka
│       └── ...
│   └── ...
└── encoding
    ├── decode.go
    └── encode.go
```
{{< /details >}}

## Package breakdown

Besides the default package breakdown (by entity type), you can put the whole code into one package by setting 
the `--package-scope all` cli arg.

{{< details "Example" >}}
```
target_dir
├── channel1.go
├── channel2.go
├── server1.go
├── server2.go
├── model1.go
├── model2.go
├── ...
├── impl
│   ├── kafka
│       └── ...
│   └── ...
└── encoding
    ├── decode.go
    └── encode.go
```
{{< /details >}}

## File breakdown

Besides the default file breakdown (by entity name), you can break down entities by the entity type by setting
the `--file-scope type` cli arg.

{{< details "Example" >}}
```
target_dir
├── channels
│   ├── channels.go
├── servers
│   ├── servers.go
├── models
│   ├── models.go
├── ...
├── impl
│   ├── kafka
│       └── ...
│   └── ...
└── encoding
    ├── decode.go
    └── encode.go
```
{{< /details >}}

Below is the combined example of the `--package-scope all` and `--file-scope type`:

{{< details "Example" >}}
```
target_dir
├── channels.go
├── servers.go
├── models.go
├── ...
├── impl
│   ├── kafka
│       └── ...
│   └── ...
└── encoding
    ├── decode.go
    └── encode.go
```
{{< /details >}}
