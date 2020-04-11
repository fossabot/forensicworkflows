// Copyright (c) 2019 Siemens AG
//
// Permission is hereby granted, free of charge, to any person obtaining a copy of
// this software and associated documentation files (the "Software"), to deal in
// the Software without restriction, including without limitation the rights to
// use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
// the Software, and to permit persons to whom the Software is furnished to do so,
// subject to the following conditions:
//
// The above copyright notice and this permission notice shall be included in all
// copies or substantial portions of the Software.
//
// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
// FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
// COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
// IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
// CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
//
// Author(s): Jonas Plum

package daggy

import (
	"fmt"

	"github.com/hashicorp/terraform/dag"
	"github.com/hashicorp/terraform/tfdiags"
	"github.com/spf13/cobra"
)

// A Task is a single element in a workflow.yml file.
type Task struct {
	Command   string                 `yaml:"command"`
	Arguments map[string]interface{} `yaml:"arguments"`
	Requires  []string               `yaml:"requires"`
}

// Workflow can be used to parse workflow.yml files.
type Workflow struct {
	Tasks map[string]Task `yaml:"tasks"`
	graph *dag.AcyclicGraph
}

// SetupGraph creates a direct acyclic graph of tasks.
func (workflow *Workflow) SetupGraph() {
	// Create the dag
	setupLogging()
	graph := dag.AcyclicGraph{}
	tasks := map[string]Task{}
	for name, task := range workflow.Tasks {
		graph.Add(name)
		tasks[name] = task
	}

	// add edges / requirements
	for name, task := range workflow.Tasks {
		for _, requirement := range task.Requires {
			graph.Connect(dag.BasicEdge(requirement, name))
		}
	}

	workflow.graph = &graph
}

// Run walks the direct acyclic graph to execute each task.
func (workflow *Workflow) Run(workingDir string, plugins map[string]*cobra.Command) error {
	w := &dag.Walker{Callback: func(v dag.Vertex) tfdiags.Diagnostics {
		task := workflow.Tasks[v.(string)]

		if plugin, ok := plugins[task.Command]; ok {
			err := workflow.runTask(plugin, task)
			if err != nil {
				return tfdiags.Diagnostics{tfdiags.Sourceless(tfdiags.Error, fmt.Sprint(v.(string)), err.Error())}
			}
			return nil
		}
		return tfdiags.Diagnostics{tfdiags.Sourceless(tfdiags.Error, fmt.Sprint(v.(string)), "command not found")}
	}}
	w.Update(workflow.graph)
	return w.Wait().Err()
}

func (workflow *Workflow) runTask(cmd *cobra.Command, task Task) error {
	args := []string{}
	for flag, value := range task.Arguments {
		args = append(args, "--"+flag, fmt.Sprint(value))
	}
	cmd.SetArgs(args)
	return cmd.Execute()
}
