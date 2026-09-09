package main

import (
	"reflect"
	"testing"

	"github.com/sensu/sensu-plugin-sdk/sensu"
)

func TestUniqueSliceElements(t *testing.T) {
	t.Run("strings", func(t *testing.T) {
		tests := []struct {
			name  string
			input []string
			want  []string
		}{
			{
				name:  "empty slice",
				input: []string{},
				want:  []string{},
			},
			{
				name:  "nil slice",
				input: nil,
				want:  []string{},
			},
			{
				name:  "no duplicates",
				input: []string{"a", "b", "c"},
				want:  []string{"a", "b", "c"},
			},
			{
				name:  "all duplicates",
				input: []string{"a", "a", "a", "a"},
				want:  []string{"a"},
			},
			{
				name:  "duplicates interleaved",
				input: []string{"a", "b", "a", "c", "b", "d", "a"},
				want:  []string{"a", "b", "c", "d"},
			},
			{
				name:  "order preservation: first occurrence order retained",
				input: []string{"z", "y", "z", "x", "y", "w"},
				want:  []string{"z", "y", "x", "w"},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := UniqueSliceElements(tt.input)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UniqueSliceElements(%v) = %v, want %v", tt.input, got, tt.want)
				}
			})
		}
	})

	t.Run("ints", func(t *testing.T) {
		tests := []struct {
			name  string
			input []int
			want  []int
		}{
			{
				name:  "empty slice",
				input: []int{},
				want:  []int{},
			},
			{
				name:  "no duplicates",
				input: []int{3, 1, 2},
				want:  []int{3, 1, 2},
			},
			{
				name:  "duplicates interleaved preserves first-occurrence order",
				input: []int{5, 2, 5, 1, 2, 3, 1},
				want:  []int{5, 2, 1, 3},
			},
		}

		for _, tt := range tests {
			t.Run(tt.name, func(t *testing.T) {
				got := UniqueSliceElements(tt.input)
				if !reflect.DeepEqual(got, tt.want) {
					t.Errorf("UniqueSliceElements(%v) = %v, want %v", tt.input, got, tt.want)
				}
			})
		}
	})
}

// savePlugin returns a copy of the package-level plugin config, so that a
// subtest can mutate the global and have it restored afterwards. checkArgs
// reads package-level mutable state (plugin.Backends / plugin.Servers), so
// tests must not leak state between subtests. These subtests intentionally
// do not run with t.Parallel() since they share that global state.
func savePlugin() Config {
	return plugin
}

func restorePlugin(saved Config) {
	plugin = saved
}

func TestCheckArgs(t *testing.T) {
	tests := []struct {
		name      string
		backends  bool
		servers   bool
		wantState int
	}{
		{
			name:      "neither backends nor servers set",
			backends:  false,
			servers:   false,
			wantState: sensu.CheckStateOK,
		},
		{
			name:      "only backends set",
			backends:  true,
			servers:   false,
			wantState: sensu.CheckStateOK,
		},
		{
			name:      "only servers set",
			backends:  false,
			servers:   true,
			wantState: sensu.CheckStateOK,
		},
		{
			name:      "both backends and servers set is mutually exclusive",
			backends:  true,
			servers:   true,
			wantState: sensu.CheckStateUnknown,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			saved := savePlugin()
			t.Cleanup(func() { restorePlugin(saved) })

			plugin.Backends = tt.backends
			plugin.Servers = tt.servers

			gotState, err := checkArgs(nil)
			if err != nil {
				t.Fatalf("checkArgs() error = %v, want nil", err)
			}
			if gotState != tt.wantState {
				t.Errorf("checkArgs() state = %d, want %d", gotState, tt.wantState)
			}
		})
	}
}

// optionField extracts a named field (e.g. "Argument" or "Value") from a
// sensu.ConfigOption. The concrete types behind the interface are generic
// instantiations of sensu.PluginConfigOption[T] / sensu.SlicePluginConfigOption[T],
// which share field names but not a common accessor method, so reflection is
// used to read them generically.
func optionField(t *testing.T, opt sensu.ConfigOption, field string) reflect.Value {
	t.Helper()
	v := reflect.ValueOf(opt)
	if v.Kind() == reflect.Pointer {
		v = v.Elem()
	}
	f := v.FieldByName(field)
	if !f.IsValid() {
		t.Fatalf("option %#v has no field %q", opt, field)
	}
	return f
}

func TestOptionsWiring(t *testing.T) {
	wantArguments := map[string]bool{
		"socket":        true,
		"backends":      true,
		"backend":       true,
		"server":        true,
		"servers":       true,
		"check-missing": true,
		"list":          true,
	}

	if len(options) != len(wantArguments) {
		t.Fatalf("len(options) = %d, want %d", len(options), len(wantArguments))
	}

	seenArguments := make(map[string]bool, len(options))

	for i, opt := range options {
		argField := optionField(t, opt, "Argument")
		if argField.Kind() != reflect.String {
			t.Fatalf("options[%d].Argument is not a string field", i)
		}
		argument := argField.String()

		if argument == "" {
			t.Errorf("options[%d] has an empty Argument", i)
		}

		if seenArguments[argument] {
			t.Errorf("options[%d] has duplicate Argument %q", i, argument)
		}
		seenArguments[argument] = true

		if !wantArguments[argument] {
			t.Errorf("options[%d] has unexpected Argument %q", i, argument)
		}

		valueField := optionField(t, opt, "Value")
		if valueField.Kind() != reflect.Pointer || valueField.IsNil() {
			t.Errorf("options[%d] (Argument %q) has a nil Value", i, argument)
		}
	}

	for arg := range wantArguments {
		if !seenArguments[arg] {
			t.Errorf("expected an option with Argument %q, but none was found", arg)
		}
	}
}

// TestPluginName documents a known bug rather than asserting it: main.go sets
// PluginConfig.Name to "check-haproxy-endpoint " with a trailing space, which
// leaks into usage/help output and the reported plugin name. Fixing it
// requires editing main.go, which is out of scope for this change (test-only
// PR). This test is skipped until that fix lands; it asserts the CORRECT
// value so it starts passing (and stays honest) the moment the trailing
// space is removed from main.go.
func TestPluginName(t *testing.T) {
	t.Skip("blocked on main.go fix: PluginConfig.Name currently has a trailing space (\"check-haproxy-endpoint \"); tracked separately, see PR description")

	want := "check-haproxy-endpoint"
	if plugin.Name != want {
		t.Errorf("plugin.Name = %q, want %q", plugin.Name, want)
	}
}
