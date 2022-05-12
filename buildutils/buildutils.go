// Copyright 2022 OnMetal authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package buildutils

import (
	"bytes"
	"fmt"
	"os/exec"
	"strings"
)

// Builder is a builder for go code.
type Builder struct {
	dir  string
	tidy bool
}

func (b *Builder) execCommand(name string, args ...string) error {
	var buf bytes.Buffer
	cmd := exec.Command(name, args...)
	cmd.Stdout = &buf
	cmd.Stderr = &buf
	cmd.Dir = b.dir

	cmdString := strings.Join(append([]string{name}, args...), " ")
	if err := cmd.Run(); err != nil {
		err = fmt.Errorf("error running %s: %w", cmdString, err)
		if output := buf.String(); output != "" {
			return fmt.Errorf("%w, output:\n%s", err, output)
		}
		return err
	}
	return nil
}

// Build runs `go build` with the target output and name.
// If BuilderOptions.Tidy was set, it runs `go mod tidy` beforehand.
func (b *Builder) Build(name, filename string) error {
	if b.tidy {
		if err := b.execCommand("go", "mod", "tidy"); err != nil {
			return fmt.Errorf("error tidying: %w", err)
		}
	}

	if err := b.execCommand("go", "build", "-o", filename, name); err != nil {
		return fmt.Errorf("error building: %w", err)
	}
	return nil
}

// BuilderOptions are options to create a builder with.
type BuilderOptions struct {
	// Dir is the working directory for the Builder.
	Dir string
	// Tidy can specify whether the builder should run `go mod tidy` before building.
	Tidy bool
}

// NewBuilder creates a new Builder with the given options.
func NewBuilder(opts BuilderOptions) *Builder {
	return &Builder{
		dir:  opts.Dir,
		tidy: opts.Tidy,
	}
}
