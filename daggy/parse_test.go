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

package daggy

import (
	"log"
	"reflect"
	"testing"
)

func TestParse(t *testing.T) {
	workflow := &Workflow{
		Tasks: map[string]Task{
			"create":      {Type: "bash", Requires: []string{"rm", "cwd"}, Command: "echo \"test\" > foo"},
			"cwd":         {Type: "bash", Requires: []string(nil), Command: "pwd"},
			"docker":      {Type: "docker", Requires: []string(nil), Image: "alpine", Command: "echo forensicreports"},
			"dockerfalse": {Type: "dockerfile", Requires: []string(nil), Image: "", Dockerfile: "jq", Command: "echo Dockerfile"},
			"false":       {Type: "bash", Requires: []string(nil), Command: "false"},
			"hello":       {Type: "plugin", Requires: []string{"cwd"}, Command: "hello.exe"},
			"plugin":      {Type: "plugin", Requires: []string{"hello", "cwd"}, Command: "example"},
			"read":        {Type: "bash", Requires: []string{"create"}, Command: "cat foo"},
			"rm":          {Type: "bash", Requires: []string(nil), Command: "rm -rf foo || true"},
			"script":      {Type: "plugin", Requires: []string{"cwd"}, Command: "pyexample"},
			"true":        {Type: "bash", Requires: []string{"false"}, Command: "true"},
		},
		// Arguments: map[string]string{"docker-server": "test.com"},
	}

	type args struct {
		workflowFile string
	}
	tests := []struct {
		name    string
		args    args
		want    *Workflow
		wantErr bool
	}{
		{"Parse example-workflow.yml", args{"../test/example-workflow.yml"}, workflow, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := Parse(tt.args.workflowFile)
			if (err != nil) != tt.wantErr {
				t.Errorf("Parse() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got == nil {
				t.Errorf("Parse() got = %#v, want %v", got, tt.want)
				return
			}

			got.graph = nil

			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Parse() got = %#v, want %v", got, tt.want)
			}
		})
	}
}

func Test_setupLogging(t *testing.T) {
	setupLogging()
	log.Print("test")
}
