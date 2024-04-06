---
title: "Code reuse"
weight: 340
# bookFlatSection: false
# bookToc: true
# bookHidden: false
# bookCollapseSection: false
# bookComments: false
# bookSearchExclude: false
---

# Code reuse

While generating the code you may want to reuse the code you already have. There are two ways to do that.

The first way is applicable when you already have the module generated before by `go-asyncapi` and want
to reuse it as a whole. For that, you can add the `--reuse-*` flags. They tell the generator to import the types 
with the same names from the specified modules instead of generating. For example:

* `--reuse-models-module github.com/foo/bar/baz` imports the models from the **github.com/foo/bar/baz** module.
* `--reuse-models-module mymodels` imports the models from the **mymodels** module in the same package.

Another way is more precise, but applicable only for models for now. The `x-go-type` extra field in a model definition 
prevents the generator from generating the type for this model and uses the specified type instead. See 
[model article]({{< relref "/docs/code-structure/model#x-go-type" >}}).