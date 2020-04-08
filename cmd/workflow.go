// Copyright (c) 2020 Siemens AG
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

package cmd

import (
	"fmt"
	"github.com/forensicanalysis/forensicworkflows/daggy"
	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"log"
	"os"
	"path"
	"path/filepath"
)

// Process is a subcommand to run a forens
func Workflow() *cobra.Command {
	command := &cobra.Command{
		Use:   "workflow",
		Short: "Run a workflow on the forensicstore",
		Long: `process can run parallel workflows locally. Those workflows are a directed acyclic graph of tasks.
Those tasks can be defined to be run on the system itself or in a containerized way.`,
		Args: func(cmd *cobra.Command, args []string) error {
			if len(args) < 1 {
				return errors.New("requires at least one store")
			}
			for _, arg := range args {
				if _, err := os.Stat(arg); os.IsNotExist(err) {
					return errors.Wrap(os.ErrNotExist, arg)
				}
			}
			return cmd.MarkFlagRequired("workflow")
		},
		Run: func(cmd *cobra.Command, args []string) {
			// parse workflow yaml
			workflowFile := cmd.Flags().Lookup("workflow").Value.String()
			if _, err := os.Stat(workflowFile); os.IsNotExist(err) {
				log.Fatal(errors.Wrap(os.ErrNotExist, workflowFile))
			}
			workflow, err := daggy.Parse(workflowFile)
			if err != nil {
				log.Fatal("parsing failed: ", err)
			}

			arguments := getArguments(cmd)
			fmt.Println(workflow, arguments)
			// tasksFunc(workflow, subcommands.Commands, "process", args, arguments)
		},
	}
	command.Flags().String("workflow", "", "workflow definition file")
	return command
}

func tasksFunc(workflow *daggy.Workflow, plugins map[string]*cobra.Command, processDir string, stores []string, arguments daggy.Arguments) {
	workflow.SetupGraph()

	// unpack scripts
	err := unpack()
	if err != nil {
		log.Fatal("unpacking error: ", err)
	}

	for _, store := range stores {
		// get store path
		storePath, err := filepath.Abs(store)
		if err != nil {
			log.Println("abs: ", err)
		}

		// run workflow
		err = workflow.Run(storePath, path.Join(processDir), plugins, arguments)
		if err != nil {
			log.Println("processing errors: ", err)
		}
	}
}

func getArguments(cmd *cobra.Command) daggy.Arguments {
	arguments := daggy.Arguments{}
	for name, unknownFlags := range cmd.Flags().UnknownFlags {
		for _, unknownFlag := range unknownFlags {
			arguments[name] = unknownFlag.Value
		}
	}
	return arguments
}
