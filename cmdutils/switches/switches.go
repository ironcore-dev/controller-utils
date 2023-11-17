// SPDX-FileCopyrightText: 2023 SAP SE or an SAP affiliate company and IronCore contributors
// SPDX-License-Identifier: Apache-2.0

// Package switches provides new type that implements flag.Value interface -- Switches.
// It can be used for enabling/disabling controllers/webhooks in your controller manager.
package switches

import (
	"encoding/csv"
	"fmt"
	"sort"
	"strings"

	"k8s.io/apimachinery/pkg/util/sets"
)

const (
	All = "*"

	disablePrefix = "-"
)

type Switches struct {
	defaults map[string]bool
	settings map[string]bool
}

// New creates an instance of Switches and returns the pointer to it
func New(settings ...string) *Switches {
	s := Make(settings...)
	return &s
}

// Make creates an instance of Switches
// Same as New but returns copy of a struct, not a pointer
func Make(settings ...string) Switches {
	s := Switches{
		defaults: make(map[string]bool),
		settings: make(map[string]bool),
	}

	s.defaults = s.prepareSettings(settings)
	return s
}

// Disable prepends disablePrefix prefix to an item name
func Disable(name string) string {
	return disablePrefix + name
}

func (s *Switches) String() string {
	var res string

	vals := make([]string, 0, len(s.defaults))
	for v := range s.defaults {
		vals = append(vals, v)
	}

	sort.Strings(vals)
	for _, v := range vals {
		if res != "" {
			res += ","
		}

		if s.settings[v] {
			res += v
		} else {
			res += "-" + v
		}
	}

	return res
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
			if _, ok := s.defaults[trimmed]; trimmed != All && !ok {
				return fmt.Errorf("unknown item: %s", trimmed)
			}
		}
	} else {
		settings = []string{""}
	}

	s.settings = s.prepareSettings(settings)
	return nil
}

// Enabled checks if item is enabled
func (s *Switches) Enabled(name string) bool {
	return s.settings[name]
}

// AllEnabled checks whether all switches with the given names are enabled.
func (s *Switches) AllEnabled(names ...string) bool {
	for _, name := range names {
		if !s.settings[name] {
			return false
		}
	}
	return true
}

// AnyEnabled checks whether any switch of the given names is enabled.
func (s *Switches) AnyEnabled(names ...string) bool {
	for _, name := range names {
		if s.settings[name] {
			return true
		}
	}
	return false
}

// All returns names of all items set in settings
func (s *Switches) All() sets.Set[string] {
	return sets.KeySet(s.defaults)
}

// Active returns names of all active items
func (s *Switches) Active() sets.Set[string] {
	names := sets.New[string]()
	for k, enabled := range s.settings {
		if enabled {
			names.Insert(k)
		}
	}

	return names
}

// Values returns the switches with their values.
func (s *Switches) Values() map[string]bool {
	res := make(map[string]bool, len(s.defaults))
	for key := range s.defaults {
		res[key] = s.settings[key]
	}
	return res
}

// EnabledByDefault returns names of all enabled items
func (s *Switches) EnabledByDefault() sets.Set[string] {
	names := sets.New[string]()
	for k, enabled := range s.defaults {
		if enabled {
			names.Insert(k)
		}
	}

	return names
}

// DisabledByDefault returns names of all disabled items
func (s *Switches) DisabledByDefault() sets.Set[string] {
	names := sets.New[string]()
	for k, enabled := range s.defaults {
		if !enabled {
			names.Insert(k)
		}
	}

	return names
}

func (s *Switches) Type() string {
	return "strings"
}

func (s *Switches) prepareSettings(settings []string) (res map[string]bool) {
	res = make(map[string]bool)

	if len(settings) == 1 && settings[0] == "" {
		return
	}

	for _, v := range settings {
		if v == All {
			for k, v := range s.defaults {
				res[k] = v
			}
			break
		}
	}

	for _, v := range settings {
		if v == All {
			continue
		}
		res[strings.TrimPrefix(v, disablePrefix)] = !strings.HasPrefix(v, disablePrefix)
	}

	return
}
