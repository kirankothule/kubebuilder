// Copyright 2019 The Kubernetes Authors.
// SPDX-License-Identifier: Apache-2.0

package commands_test

import (
	"bytes"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"sigs.k8s.io/kustomize/cmd/config/internal/commands"
)

const (
	flagsInput = `kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  spec:
    template:
      spec:
        containers:
        - name: nginx
          image: nginx
- apiVersion: apps/v1
  kind: Service
  spec: {}
functionConfig:
  kind: Foo
  spec:
    a: b
    c: d
    e: f
  items:
  - 1
  - 3
  - 2
  - 4
`

	resourceInput = `apiVersion: config.kubernetes.io/v1alpha1
kind: ResourceList
items:
- apiVersion: apps/v1
  kind: Deployment
  spec:
    template:
      spec:
        containers:
        - name: nginx
          image: nginx
- apiVersion: apps/v1
  kind: Service
  spec: {}
functionConfig:
  kind: Foo
`

	resourceOutput = `apiVersion: v1
kind: List
items:
- apiVersion: apps/v1
  kind: Deployment
  spec:
    template:
      spec:
        containers:
        - name: nginx
          image: nginx
- apiVersion: apps/v1
  kind: Service
  spec: {}
`
)

func TestXArgs_flags(t *testing.T) {
	c := commands.GetXArgsRunner()
	c.Command.SetIn(bytes.NewBufferString(flagsInput))
	out := &bytes.Buffer{}
	c.Command.SetOut(out)
	c.Command.SetArgs([]string{"--", "echo"})

	c.Args = []string{"--", "echo"}
	if !assert.NoError(t, c.Command.Execute()) {
		t.FailNow()
	}
	assert.Equal(t, strings.TrimSpace(`--a=b --c=d --e=f 1 3 2 4`), strings.TrimSpace(out.String()))
}

func TestXArgs_input(t *testing.T) {
	c := commands.GetXArgsRunner()
	c.Command.SetIn(bytes.NewBufferString(resourceInput))
	out := &bytes.Buffer{}
	c.Command.SetOut(out)
	c.Command.SetArgs([]string{"--", "cat"})

	c.Args = []string{"--", "cat"}
	if !assert.NoError(t, c.Command.Execute()) {
		t.FailNow()
	}
	assert.Equal(t, resourceOutput, out.String())
}

func TestCmd_env(t *testing.T) {
	c := commands.GetXArgsRunner()
	c.Command.SetIn(bytes.NewBufferString(flagsInput))
	out := &bytes.Buffer{}
	c.Command.SetOut(out)
	c.Command.SetArgs([]string{"--env-only", "--", "env"})

	c.Args = []string{"--", "env"}
	if !assert.NoError(t, c.Command.Execute()) {
		t.FailNow()
	}
	assert.Contains(t, out.String(), "\nA=b\nC=d\nE=f\n")
}
