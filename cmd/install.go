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
	"errors"
	"fmt"
	"github.com/markbates/pkger"
	"github.com/spf13/cobra"
	"io"
	"log"
	"os"
	"path/filepath"
	"strings"
)

// Install is a subcommand to run a forens
func Install() *cobra.Command {
	var force *bool

	command := &cobra.Command{
		Use:   "install",
		Short: "Setup required artifacts",
		RunE: func(cmd *cobra.Command, args []string) error {
			configDir, err := os.UserConfigDir()
			if err != nil {
				return err
			}

			appDir := filepath.Join(configDir, "forensicstore") // TODO: appname

			info, err := os.Stat(appDir)
			if err != nil && !os.IsNotExist(err) {
				return err
			}

			if !info.IsDir() {
				return fmt.Errorf("%s is not a directory", appDir)
			}
			if os.IsNotExist(err) || *force {
				return unpack()
			}
			return nil // fmt.Errorf("%s already exists, use --force to recreate", appDir)
		},
	}
	command.Flags().BoolVarP(force, "force", "f", false, "workflow definition file")
	return command
}

func unpack() error {
	cacheDir, err := os.UserConfigDir()
	if err != nil {
		return err
	}

	forensicstoreDir := filepath.Join(cacheDir, "forensicstore") // TODO: appname
	scriptsDir := filepath.Join(forensicstoreDir, "scripts")

	_ = os.RemoveAll(scriptsDir)

	log.Printf("unpack to %s\n", forensicstoreDir)

	err = pkger.Walk("/scripts", func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		parts := strings.SplitN(path, ":", 2)
		if len(parts) != 2 {
			return errors.New("could not split path")
		}

		if info.IsDir() {
			return os.MkdirAll(filepath.Join(forensicstoreDir, parts[1]), 0700)
		}

		// Copy file
		err = os.MkdirAll(filepath.Join(forensicstoreDir, filepath.Dir(parts[1])), 0700)
		if err != nil {
			return err
		}
		srcFile, err := pkger.Open(parts[1])
		if err != nil {
			return err
		}
		dstFile, err := os.Create(filepath.Join(forensicstoreDir, parts[1]))
		if err != nil {
			return err
		}
		_, err = io.Copy(dstFile, srcFile)
		return err
	})

	return err
}
