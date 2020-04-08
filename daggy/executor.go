/*
 * Copyright (c) 2020 Siemens AG
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy of
 * this software and associated documentation files (the "Software"), to deal in
 * the Software without restriction, including without limitation the rights to
 * use, copy, modify, merge, publish, distribute, sublicense, and/or sell copies of
 * the Software, and to permit persons to whom the Software is furnished to do so,
 * subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in all
 * copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY, FITNESS
 * FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE AUTHORS OR
 * COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER LIABILITY, WHETHER
 * IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM, OUT OF OR IN
 * CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN THE SOFTWARE.
 *
 * Author(s): Jonas Plum
 */

package daggy

import (
	"context"
	"encoding/json"
	"io/ioutil"
	"log"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
	"github.com/spf13/cobra"
)

const appName = "forensicstore"

func DockerCommands() []*cobra.Command {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return nil
	}

	options := types.ImageListOptions{All: true}
	imageSummaries, err := cli.ImageList(ctx, options)
	if err != nil {
		return nil
	}

	var commands []*cobra.Command
	for _, imageSummary := range imageSummaries {
		for _, name := range imageSummary.RepoTags {
			idx := strings.LastIndex(name, "/")
			if strings.HasPrefix(name[idx+1:], appName+"-") {
				commands = append(commands, dockerCommand(name))
			}
		}
	}

	return commands
}

func dockerCommand(image string) *cobra.Command {
	return &cobra.Command{Use: image}
}

func ScriptCommands() []*cobra.Command {
	dir, _ := os.UserConfigDir()
	scriptDir := filepath.Join(dir, "forensicstore", "scripts")
	infos, _ := ioutil.ReadDir(scriptDir)

	var commands []*cobra.Command
	for _, info := range infos {
		if info.Mode().IsRegular() && strings.HasPrefix(info.Name(), appName+"-") {
			commands = append(commands, scriptCommand(filepath.Join(scriptDir, info.Name())))
		}
	}

	return commands
}

func scriptCommand(path string) *cobra.Command {
	var cmd cobra.Command

	sh := exec.Command("sh", "-c", filepath.ToSlash(path)+" info")
	sh.Stderr = log.Writer()

	out, err := sh.Output()
	if err != nil {
		log.Println(path, err)
	} else {
		err = json.Unmarshal(out, &cmd)
		if err != nil {
			log.Println(err)
		}
	}

	if cmd.Use == "" {
		cmd.Use = filepath.Base(path)
	}
	cmd.RunE = func(cmd *cobra.Command, args []string) error {
		return exec.Command(path, args...).Run()
	}
	return &cmd
}
