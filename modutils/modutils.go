// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

package modutils

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/fs"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"time"

	"github.com/ironcore-dev/controller-utils/buildutils"
)

// Executor is an executor for go.mod-related operations.
type Executor struct {
	dir string
}

// ExecutorOptions are options to create an executor with.
type ExecutorOptions struct {
	// Dir is the working directory the executor runs in.
	Dir string
}

// NewExecutor creates a new Executor with the given ExecutorOptions.
func NewExecutor(opts ExecutorOptions) *Executor {
	return &Executor{dir: opts.Dir}
}

// Module is a module read from a go.mod file and its dependencies.
type Module struct {
	Path      string
	Version   string
	Replace   *Module
	Time      time.Time
	Indirect  bool
	Main      bool
	Dir       string
	GoMod     string
	GoVersion string
}

func (e *Executor) execGoModList(pattern string) ([]Module, error) {
	var (
		stdout bytes.Buffer
		stderr bytes.Buffer
	)
	cmd := exec.Command("go", "list", "-json", "-m", pattern)
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	cmd.Dir = e.dir
	if err := cmd.Run(); err != nil {
		return nil, fmt.Errorf("error executing go mod list -m %s:\n\n%s", pattern, stderr.String())
	}

	var modules []Module
	dec := json.NewDecoder(&stdout)
	for dec.More() {
		var mod Module
		if err := dec.Decode(&mod); err != nil {
			return nil, fmt.Errorf("error decoding module: %w", err)
		}

		modules = append(modules, mod)
	}
	return modules, nil
}

// ListE lists all modules of the current module.
func (e *Executor) ListE() ([]Module, error) {
	return e.execGoModList("all")
}

// List lists all modules of the current module. It panics if an error occurs.
func (e *Executor) List() []Module {
	mods, err := e.ListE()
	if err != nil {
		panic(err)
	}
	return mods
}

// GetE gets the module with the specified name.
func (e *Executor) GetE(name string) (*Module, error) {
	mods, err := e.execGoModList(name)
	if err != nil {
		return nil, fmt.Errorf("error loading modules: %w", err)
	}

	for _, mod := range mods {
		if mod.Path != name {
			continue
		}

		mod := mod
		return &mod, nil
	}
	return nil, fmt.Errorf("module %q not found", name)
}

// Get gets the module with the specified name.
// It panics if an error occurs.
func (e *Executor) Get(name string) *Module {
	mod, err := e.GetE(name)
	if err != nil {
		panic(err)
	}
	return mod
}

// DirE gets the on-disk location of the specified module, optionally joining the parts to the directory.
func (e *Executor) DirE(name string, parts ...string) (string, error) {
	mod, err := e.GetE(name)
	if err != nil {
		return "", fmt.Errorf("error getting module: %w", err)
	}
	if mod.Dir == "" {
		return "", fmt.Errorf("module %s does not specify a directory", name)
	}

	if len(parts) == 0 {
		return mod.Dir, nil
	}
	return filepath.Join(append([]string{mod.Dir}, parts...)...), nil
}

// Dir gets the on-disk location of the specified module, optionally joining the parts to the directory.
// It panics if an error occurs.
func (e *Executor) Dir(name string, parts ...string) string {
	dir, err := e.DirE(name, parts...)
	if err != nil {
		panic(err)
	}
	return dir
}

func (e *Executor) copy(src, dst string) error {
	return filepath.Walk(src, func(path string, info fs.FileInfo, err error) error {
		if err != nil {
			return err
		}

		relPath, err := filepath.Rel(src, path)
		if err != nil {
			return err
		}

		dstPath := filepath.Join(dst, relPath)
		if info.IsDir() {
			return os.MkdirAll(dstPath, 0777)
		} else {
			return copyFile(path, dstPath)
		}
	})
}

func copyFile(srcFilename, dstFilename string) error {
	srcFile, err := os.Open(srcFilename)
	if err != nil {
		return fmt.Errorf("error opening source file %s: %w", srcFilename, err)
	}
	defer func() { _ = srcFile.Close() }()

	srcStat, err := srcFile.Stat()
	if err != nil {
		return fmt.Errorf("error stat-ing source file %s: %w", srcFilename, err)
	}

	dstFile, err := os.OpenFile(dstFilename, os.O_RDWR|os.O_CREATE|os.O_TRUNC, srcStat.Mode()|0666)
	if err != nil {
		return fmt.Errorf("error creating destination file %s: %w", dstFilename, err)
	}

	if _, err := io.Copy(dstFile, srcFile); err != nil {
		_ = dstFile.Close()
		_ = os.Remove(dstFilename)
		return fmt.Errorf("error copying from source file %s to destination file %s: %w", srcFilename, dstFilename, err)
	}

	if err := dstFile.Close(); err != nil {
		_ = os.Remove(dstFilename)
		return fmt.Errorf("error closing destination file %s: %w", dstFilename, err)
	}
	return nil
}

// BuildE builds the specified module to the target filename, optionally taking sub-paths in the target module.
func (e *Executor) BuildE(filename, name string, parts ...string) error {
	dir, err := e.DirE(name)
	if err != nil {
		return fmt.Errorf("error getting directory of %s: %w", name, err)
	}

	target := "."
	if len(parts) > 0 {
		target = "./" + path.Join(parts...)
	}

	if err := e.build(name, dir, target, filename); err != nil {
		return fmt.Errorf("error building %s: %w", name, err)
	}
	return nil
}

// Build builds the specified module to the target filename, optionally taking sub-paths in the target module.
// It panics if an error occurs.
func (e *Executor) Build(filename, name string, parts ...string) {
	if err := e.BuildE(filename, name, parts...); err != nil {
		panic(err)
	}
}

func (e *Executor) build(name, dir, target, filename string) error {
	buildDir, err := os.MkdirTemp("", "build-")
	if err != nil {
		return fmt.Errorf("error creating temp directory: %w", err)
	}
	defer func() { _ = os.RemoveAll(buildDir) }()

	if err := e.copy(dir, buildDir); err != nil {
		return fmt.Errorf("error copying module to build directory: %w", err)
	}

	bldr := buildutils.NewBuilder(buildutils.BuilderOptions{
		Dir:  buildDir,
		Tidy: true,
	})

	if err := bldr.Build(target, filename); err != nil {
		return fmt.Errorf("error building %s (target %s): %w", name, target, err)
	}
	return nil
}
