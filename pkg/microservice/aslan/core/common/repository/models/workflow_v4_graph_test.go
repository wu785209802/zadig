package models

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gopkg.in/yaml.v3"
)

func TestWorkflowV4_HashIgnoresGraph(t *testing.T) {
	w1 := &WorkflowV4{Name: "wf", Project: "p", Stages: []*WorkflowStage{}}
	w2 := &WorkflowV4{
		Name:    "wf",
		Project: "p",
		Stages:  []*WorkflowStage{},
		Graph: map[string]interface{}{
			"version": float64(1),
			"nodes":   []interface{}{},
			"edges":   []interface{}{},
		},
	}
	assert.Equal(t, w1.CalculateHash(), w2.CalculateHash())
}

func TestWorkflowV4_GraphYAMLRoundTrip(t *testing.T) {
	raw := []byte(`
name: wf
project: p
stages: []
graph:
  version: 1
  nodes:
    - id: n1
      type: zadig-build
      name: build
      position:
        x: 1
        y: 2
      spec: {}
      inputs: []
      outputs: []
  edges: []
`)
	var w WorkflowV4
	require.NoError(t, yaml.Unmarshal(raw, &w))
	require.NotNil(t, w.Graph)
	assert.EqualValues(t, 1, w.Graph["version"])

	b, err := json.Marshal(w)
	require.NoError(t, err)
	var again WorkflowV4
	require.NoError(t, json.Unmarshal(b, &again))
	require.NotNil(t, again.Graph)
	assert.EqualValues(t, 1, again.Graph["version"])
}
