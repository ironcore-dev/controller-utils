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
	"encoding/csv"
	"fmt"
	"strings"

	"sigs.k8s.io/kustomize/kyaml/sets"
)

const (
	DefaultValue = "*"

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

// Disable prepends disablePrefix prefix to an item name
func Disable(name string) string {
	return disablePrefix + name
}

func (s *Switches) String() string {
	return fmt.Sprintf("%v", s.settings)
}

func (s *Switches) Set(val string) error {
	var (
		err      error
		settings []string
	)

	if val != "" {
		stringReader := strings.NewReader(val)
		csvReader := csv.NewReader(stringReader)

		settings, err = csvReader.Read()
		if err != nil {
			return fmt.Errorf("failed to set switches value: %w", err)
		}

		// Validate that all specified controllers are known
		for _, v := range settings {
			trimmed := strings.TrimPrefix(v, disablePrefix)
			if _, ok := s.settings[trimmed]; trimmed != DefaultValue && !ok {
				return fmt.Errorf("unknown item: %s", trimmed)
			}
		}
	} else {
		settings = []string{""}
	}

	s.setSettings(settings)
	return nil
}

// Enabled checks if item is enabled
func (s *Switches) Enabled(name string) bool {
	return s.settings[name]
}

// All returns names of all items set in settings
func (s *Switches) All() sets.String {
	names := make(sets.String, len(s.settings))
	for k := range s.settings {
		names.Insert(k)
	}

	return names
}

// DisabledByDefault returns names of all disabled items
func (s *Switches) DisabledByDefault() sets.String {
	names := make(sets.String)
	for k, enabled := range s.settings {
		if !enabled {
			names.Insert(k)
		}
	}

	return names
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
		if v == DefaultValue {
			isDefault = true
			break
		}
	}

	if !isDefault {
		for k := range s.settings {
			s.settings[k] = false
		}
	}

	for _, v := range settings {
		if v == DefaultValue {
			continue
		}
		s.settings[strings.TrimPrefix(v, disablePrefix)] = !strings.HasPrefix(v, disablePrefix)
	}
}
