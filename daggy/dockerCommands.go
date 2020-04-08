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
	"context"
	"fmt"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/docker/docker/client"
)

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
	var dockerUser, dockerPassword, dockerServer string
	cmd := &cobra.Command{
		Use: image,
		RunE: func(cmd *cobra.Command, args []string) error {
			var auth types.AuthConfig
			auth.Username = dockerUser
			auth.Password = dockerPassword
			auth.ServerAddress = dockerServer

			for _, url := range args {
				i := 0
				cmd.VisitParents(func(_ *cobra.Command) {
					i++
				})

				err := docker(image, os.Args[i:], true, auth, url)
				if err != nil {
					return err
				}
			}
			return nil
		},
	}
	cmd.PersistentFlags().StringVar(&dockerUser, "docker-user", "", "docker registry username")
	cmd.PersistentFlags().StringVar(&dockerPassword, "docker-password", "", "docker registry password")
	cmd.PersistentFlags().StringVar(&dockerServer, "docker-server", "", "docker registry server")
	return cmd
}

func docker(image string, args []string, pull bool, auth types.AuthConfig, storePath string) error {
	fmt.Println(image, args, storePath)
	return nil
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		return err
	}

	if pull {
		err = pullImage(ctx, cli, image, auth)
		if err != nil {
			return err
		}
	}

	workingDir := storePath

	// create directory if not exists
	_, err = os.Open(workingDir)
	if os.IsNotExist(err) {
		log.Println("creating directory", workingDir)
		err = os.MkdirAll(workingDir, os.ModePerm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	if workingDir[1] == ':' {
		workingDir = "/" + strings.ToLower(string(workingDir[0])) + filepath.ToSlash(workingDir[2:])
	}

	resp, err := createContainer(ctx, cli, image, args, workingDir)
	if err != nil {
		return err
	}

	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		return err
	}

	statusChannel, errChannel := cli.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)
	select {
	case err := <-errChannel:
		if err != nil {
			return err
		}
	case <-statusChannel:
	}

	out, err := cli.ContainerLogs(ctx, resp.ID, types.ContainerLogsOptions{ShowStdout: true})
	if err != nil {
		return err
	}

	// stdcopy.StdCopy(os.Stdout, os.Stderr, out)
	// _, err = ioutil.ReadAll(out)
	_, err = io.Copy(log.Writer(), out)
	return err
}

func createContainer(ctx context.Context, cli *client.Client, image string, args []string, workingDir string) (container.ContainerCreateCreatedBody, error) {
	mounts := []mount.Mount{{Type: mount.TypeBind, Source: workingDir, Target: "/store"}}

	/*
		// add transit dir if import or export
		transitPath := arguments.Get("file")
		if transitPath != "" {
			transitDir, transitFile := filepath.Split(transitPath)
			if transitDir[1] == ':' {
				transitDir = "/" + strings.ToLower(string(transitDir[0])) + filepath.ToSlash(transitDir[2:])
			}
			mounts = append(mounts, mount.Mount{Type: mount.TypeBind, Source: transitDir, Target: "/transit"})
			cmd = append(cmd, "--file", transitFile)
		}
	*/


	resp, err := cli.ContainerCreate(
		ctx,
		&container.Config{Image: image, Cmd: args, Tty: true, WorkingDir: "/store"},
		&container.HostConfig{Mounts: mounts},
		nil,
		"",
	)
	if err != nil {
		return container.ContainerCreateCreatedBody{}, err
	}
	return resp, nil
}

func pullImage(ctx context.Context, cli *client.Client, image string, auth types.AuthConfig) error {
	body, err := cli.RegistryLogin(ctx, auth)
	if err != nil {
		return err
	}
	log.Println("login", body)

	reader, err := cli.ImagePull(ctx, image, types.ImagePullOptions{})
	if err != nil {
		return err
	}
	_, err = io.Copy(os.Stderr, reader)
	if err != nil {
		return err
	}
	return nil
}
