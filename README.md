# Tutorials and information around {{PROJECT}}

## Overview of this tutorial

{{Overview}}

## Table of Contents

Tutorial listing

1. [Prereqs](#prerequisites)
2. [Tutorial Breakouts](#tutorials)
3. [Reference Docs](#reference-docs)

---

## Prerequisites

- OCP
- CLI

---

## Tutorials

Building an operator

This will generally scaffold the project

```bash
operator-sdk init \
  --domain=workshop.io \
  --repo=github.com/axodevelopment/demo-operators \
  --plugins=go/v4
```

Now we need to add the controller

```bash
operator-sdk create api \
  --group=workshop \
  --version=v1 \
  --kind=Paychex \
  --resource --controller
```

api/v1/paychex_types.go was created.  Add the lables for the paychex type that will be captured.

```bash
type PaychexSpec struct {
    Labels map[string]string `json:"labels,omitempty"`
}
```

Once all the edits to the api struct we leverage make generate command to create the CRD definition from this data.

```bash
make generate manifests
```