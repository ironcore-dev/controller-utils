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

// ModMode is the module download mode to use.
type ModMode string

const (
	// ModModeVendor causes modules to be resolved from a vendor folder.
	ModModeVendor ModMode = "vendor"
	// ModModeReadonly expect all modules to be present in the module cache for the current module.
	ModModeReadonly ModMode = "readonly"
	// ModModeMod fetches any module before building.
	ModModeMod ModMode = "mod"
)

// ApplyToBuild implements BuildOption.
func (m ModMode) ApplyToBuild(o *BuildOptions) {
	o.Mod = &m
}

// BuildOptions are options to supply for a Build.
type BuildOptions struct {
	// ForceRebuild forces rebuilding of packages that are already up-to-date.
	ForceRebuild bool
	// Mod specifies the module download mode to use.
	Mod *ModMode
}

// ApplyOptions applies the slice of BuildOption to this BuildOptions.
func (o *BuildOptions) ApplyOptions(opts []BuildOption) {
	for _, opt := range opts {
		opt.ApplyToBuild(o)
	}
}

// ApplyToBuild implements BuildOption.
func (o *BuildOptions) ApplyToBuild(o2 *BuildOptions) {
	if o.ForceRebuild {
		o2.ForceRebuild = true
	}
	if o.Mod != nil {
		o2.Mod = o.Mod
	}
}

// BuildOption are options to apply to BuildOptions.
type BuildOption interface {
	// ApplyToBuild applies the option to the BuildOptions.
	ApplyToBuild(o *BuildOptions)
}

// forceRebuild is an option to force rebuilding packages.
type forceRebuild struct{}

// ApplyToBuild implements BuildOption.
func (forceRebuild) ApplyToBuild(o *BuildOptions) {
	o.ForceRebuild = true
}

// ForceRebuild is an option to force rebuilding packages.
var ForceRebuild = forceRebuild{}

// Build runs `go build` with the target output and name.
// If BuilderOptions.Tidy was set, it runs `go mod tidy` beforehand.
func (b *Builder) Build(name, filename string, opts ...BuildOption) error {
	o := &BuildOptions{}
	o.ApplyOptions(opts)

	if b.tidy {
		if err := b.execCommand("go", "mod", "tidy"); err != nil {
			return fmt.Errorf("error tidying: %w", err)
		}
	}

	args := []string{"build", "-o", filename}

	if mod := o.Mod; mod != nil {
		args = append(args, "-mod", string(*mod))
	}
	if o.ForceRebuild {
		args = append(args, "-a")
	}

	args = append(args, name)

	if err := b.execCommand("go", args...); err != nil {
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
