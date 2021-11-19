// Copyright 2021 OnMetal authors
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

// Package switches provides new type that implements flag.Value interface -- Switches.
// It can be used for enabling/disabling controllers/webhooks in your controller manager.
package switches

import (
	"fmt"
	"strings"
)

const (
	defaultValue  = "*"
	disablePrefix = "-"
)

type Switches struct {
	settings map[string]bool
}

// New creates an instance of Switches
func New(settings []string) *Switches {
	s := &Switches{
		settings: make(map[string]bool),
	}
	s.setSettings(settings)
	return s
}

// Disable prepends disablePrefix prefix to a controller name
func Disable(name string) string {
	return disablePrefix + name
}

func (s *Switches) String() string {
	return fmt.Sprintf("%v", s.settings)
}

func (s *Switches) Set(val string) error {
	s.setSettings(strings.Split(val, ","))
	return nil
}

// Enabled checks if controller is enabled
func (s *Switches) Enabled(name string) bool {
	return s.settings[name]
}

// All returns names of all controllers set in settings
func (s *Switches) All() (names []string) {
	for k := range s.settings {
		names = append(names, k)
	}

	return
}

// DisabledByDefault returns names of all disabled controllers
func (s *Switches) DisabledByDefault() (names []string) {
	for k, enabled := range s.settings {
		if !enabled {
			names = append(names, k)
		}
	}

	return
}

func (s *Switches) Type() string {
	return "Switches"
}

func (s *Switches) setSettings(settings []string) {
	if len(settings) == 1 && settings[0] == "" {
		return
	}

	var isDefault bool
	for _, v := range settings {
		if v == defaultValue {
			isDefault = true
			break
		}
	}

	if !isDefault {
		s.settings = make(map[string]bool)
	}

	for _, v := range settings {
		if v == defaultValue {
			continue
		}
		s.settings[strings.TrimPrefix(v, disablePrefix)] = !strings.HasPrefix(v, disablePrefix)
	}
}
