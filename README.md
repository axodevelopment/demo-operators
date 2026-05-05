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


NOTE: This geneatees workshop.workshop.io/v1

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

Need to get the route crd

```bash
go get github.com/openshift/api@latest
```

Add the route dependecy in main

What is a Manager

What is a Reconciler

init()

---

Onto the Reconciler logic

---

Resource mapping of ocp version may need additional pin's

```bash
go get k8s.io/api@v0.35.1 \
       k8s.io/apimachinery@v0.35.1 \
       k8s.io/client-go@v0.35.1
```


```bash
go mod tidy
```

```bash
make manifests
```

```bash
make install
```

```bash
make run
```

new terminal deploy config/samples

```bash
apiVersion: workshop.workshop.io/v1
kind: Paychex
metadata:
  name: test1
  namespace: my-test
spec:
  labels:
    owner: operator
```

```bash
oc get all,route -n my-test
```