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
	Type      string              `yaml:"type"`
	Requires  []string            `yaml:"requires"`
	Command   string              `yaml:"command"` // shared
	Arguments map[string][]string `yaml:"with"`
}

// Workflow can be used to parse workflow.yml files.
type Workflow struct {
	Tasks      map[string]Task `yaml:"tasks"`
	graph      *dag.AcyclicGraph
	workingDir string
	pluginDir  string
	plugins    map[string]*cobra.Command
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
func (workflow *Workflow) Run(workingDir, pluginDir string, plugins map[string]*cobra.Command, arguments []string) error {
	workflow.workingDir = workingDir
	workflow.pluginDir = pluginDir
	workflow.plugins = plugins

	w := &dag.Walker{Callback: func(v dag.Vertex) tfdiags.Diagnostics {
		err := workflow.runTask(v.(string), arguments)
		if err != nil {
			return tfdiags.Diagnostics{tfdiags.Sourceless(tfdiags.Error, fmt.Sprint(v.(string)), err.Error())}
		}
		return nil
	}}
	w.Update(workflow.graph)
	return w.Wait().Err()
}

func (workflow *Workflow) runTask(taskName string, arguments []string) error {
	// task := workflow.Tasks[taskName]
	//process := cmd.Process()
	//process.SetArgs(append([]string{taskName}, arguments...))
	//return process.Execute()
	return nil
}
