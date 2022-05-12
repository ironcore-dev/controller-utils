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

package modutils

var (
	// DefaultExecutor is the default executor.
	DefaultExecutor = &Executor{}

	// ListE is an alias to DefaultExecutor.ListE.
	ListE = DefaultExecutor.ListE
	// List is an alias to DefaultExecutor.List.
	List = DefaultExecutor.List

	// GetE is an alias to DefaultExecutor.GetE.
	GetE = DefaultExecutor.GetE
	// Get is an alias to DefaultExecutor.Get.
	Get = DefaultExecutor.Get

	// GetDirE is an alias to DefaultExecutor.GetDirE.
	GetDirE = DefaultExecutor.DirE
	// GetDir is an alias to DefaultExecutor.GetDir.
	GetDir = DefaultExecutor.Dir

	// BuildE is an alias to DefaultExecutor.BuildE.
	BuildE = DefaultExecutor.BuildE
	// Build is an alias to DefaultExecutor.Build.
	Build = DefaultExecutor.Build
)
