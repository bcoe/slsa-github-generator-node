// Copyright The GOSST team.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package pkg

import (
	"fmt"
	// "os"
	"testing"

	"github.com/google/go-cmp/cmp"
	// "github.com/google/go-cmp/cmp/cmpopts"
)

func Test_isAllowedEnvVariable(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		variable string
		expected bool
	}{
		{
			name:     "BLA variable",
			variable: "BLA",
			expected: false,
		},
		{
			name:     "random variable",
			variable: "random",
			expected: false,
		},
		{
			name:     "NODE_SOMETHING variable",
			variable: "NODE_SOMETHING",
			expected: true,
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := isAllowedEnvVariable(tt.variable)
			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}

func Test_marshallList(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		variables []string
		expected  string
	}{
		{
			name:      "single arg",
			variables: []string{"--arg"},
			expected:  "WyItLWFyZyJd",
		},
		{
			name: "list args",
			variables: []string{
				"/usr/lib/google-golang/bin/go",
				"build", "-mod=vendor", "-trimpath",
				"-tags=netgo",
				"-ldflags=-X main.gitVersion=v1.2.3 -X main.gitSomething=somthg",
			},
			expected: "WyIvdXNyL2xpYi9nb29nbGUtZ29sYW5nL2Jpbi9nbyIsImJ1aWxkIiwiLW1vZD12ZW5kb3IiLCItdHJpbXBhdGgiLCItdGFncz1uZXRnbyIsIi1sZGZsYWdzPS1YIG1haW4uZ2l0VmVyc2lvbj12MS4yLjMgLVggbWFpbi5naXRTb21ldGhpbmc9c29tdGhnIl0=",
		},
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r, err := marshallList(tt.variables)
			if err != nil {
				t.Errorf("marshallList: %v", err)
			}
			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}

func Test_isAllowedArg(t *testing.T) {
	t.Parallel()

	var tests []struct {
		name     string
		argument string
		expected bool
	}

	for k := range allowedBuildArgs {
		tests = append(tests, struct {
			name     string
			argument string
			expected bool
		}{
			name:     fmt.Sprintf("%s argument", k),
			argument: k,
			expected: true,
		})

		tests = append(tests, struct {
			name     string
			argument string
			expected bool
		}{
			name:     fmt.Sprintf("%sbla argument", k),
			argument: fmt.Sprintf("%sbla", k),
			expected: true,
		})

		tests = append(tests, struct {
			name     string
			argument string
			expected bool
		}{
			name:     fmt.Sprintf("bla %s argument", k),
			argument: fmt.Sprintf("bla%s", k),
			expected: false,
		})

		tests = append(tests, struct {
			name     string
			argument string
			expected bool
		}{
			name:     fmt.Sprintf("space %s argument", k),
			argument: fmt.Sprintf(" %s", k),
			expected: false,
		})
	}
	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			r := isAllowedArg(tt.argument)
			if !cmp.Equal(r, tt.expected) {
				t.Errorf(cmp.Diff(r, tt.expected))
			}
		})
	}
}

func Test_generateOutputFilename(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		version  string
		expected struct {
			err error
			fn  string
		}
	}{
		{
			name:    "foo-pkg",
			version: "1.2.3",
			expected: struct {
				err error
				fn  string
			}{
				err: nil,
				fn:  "foo-pkg-1.2.3.tgz",
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := pkgJsonConfigFile{
				Name:    tt.name,
				Version: tt.version,
			}
			c, err := pkgJSONFromConfig(&cfg)
			if err != nil {
				t.Errorf("pkgJSONFromConfig: %v", err)
			}
			b := NodeBuildNew("node compiler", "npm", c)

			fn, err := b.generateOutputFilename()
			if !errCmp(err, tt.expected.err) {
				t.Errorf(cmp.Diff(err, tt.expected.err))
			}

			if err != nil {
				return
			}

			if fn != tt.expected.fn {
				t.Errorf(cmp.Diff(fn, tt.expected.fn))
			}
		})
	}
}

// TODO: implement arg env variables for Node builder.
/*
func Test_SetArgEnvVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		argEnv   string
		expected struct {
			err error
			env map[string]string
		}
	}{
		{
			name:   "valid arg envs",
			argEnv: "VAR1:value1, VAR2:value2",
			expected: struct {
				err error
				env map[string]string
			}{
				err: nil,
				env: map[string]string{"VAR1": "value1", "VAR2": "value2"},
			},
		},
		{
			name:   "empty arg envs",
			argEnv: "",
			expected: struct {
				err error
				env map[string]string
			}{
				err: nil,
				env: map[string]string{},
			},
		},
		{
			name:   "valid arg envs not space",
			argEnv: "VAR1:value1,VAR2:value2",
			expected: struct {
				err error
				env map[string]string
			}{
				err: nil,
				env: map[string]string{"VAR1": "value1", "VAR2": "value2"},
			},
		},
		{
			name:   "invalid arg empty 2 values",
			argEnv: "VAR1:value1,",
			expected: struct {
				err error
				env map[string]string
			}{
				err: errorInvalidEnvArgument,
			},
		},
		{
			name:   "invalid arg empty 3 values",
			argEnv: "VAR1:value1,, VAR3:value3",
			expected: struct {
				err error
				env map[string]string
			}{
				err: errorInvalidEnvArgument,
			},
		},
		{
			name:   "invalid arg uses equal",
			argEnv: "VAR1=value1",
			expected: struct {
				err error
				env map[string]string
			}{
				err: errorInvalidEnvArgument,
			},
		},
		{
			name:   "valid single arg",
			argEnv: "VAR1:value1",
			expected: struct {
				err error
				env map[string]string
			}{
				err: nil,
				env: map[string]string{"VAR1": "value1"},
			},
		},
		{
			name:   "invalid valid single arg with empty",
			argEnv: "VAR1:value1:",
			expected: struct {
				err error
				env map[string]string
			}{
				err: errorInvalidEnvArgument,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := goReleaserConfigFile{
				Version: 1,
			}
			c, err := fromConfig(&cfg)
			if err != nil {
				t.Errorf("fromConfig: %v", err)
			}
			b := NodeBuildNew("go compiler", c)

			err = b.SetArgEnvVariables(tt.argEnv)
			if !errCmp(err, tt.expected.err) {
				t.Errorf(cmp.Diff(err, tt.expected.err))
			}

			if err != nil {
				return
			}

			sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if !cmp.Equal(b.argEnv, tt.expected.env, sorted) {
				t.Errorf(cmp.Diff(b.argEnv, tt.expected.env))
			}
		})
	}
}
*/

// TODO: discuss environment variables required for Node.js builder.
/*
func Test_generateEnvVariables(t *testing.T) {
	t.Parallel()

	tests := []struct {
		env      []string
		expected struct {
			err   error
			flags []string
		}
	}{
		{
			env:    []string{"VAR1=value1", "VAR2=value2"},
			expected: struct {
				err   error
				flags []string
			}{
				err: errorEnvVariableNameNotAllowed,
			},
		},
		{
			env:    []string{"GOVAR1=value1", "GOVAR2=value2", "CGO_VAR1=val1", "CGO_VAR2=val2"},
			expected: struct {
				err   error
				flags []string
			}{
				flags: []string{
					"GOOS=windows", "GOARCH=amd64",
					"GOVAR1=value1", "GOVAR2=value2", "CGO_VAR1=val1", "CGO_VAR2=val2",
				},
				err: nil,
			},
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run("foo-pkg", func(t *testing.T) {
			t.Parallel()

			cfg := pkgJsonConfigFile{
				Name: "foo-pkg",
				Version:    "1.2.3",
			}
			c, err := pkgJSONFromConfig(&cfg)
			if err != nil {
				t.Errorf("pkgJSONFromConfig: %v", err)
			}
			b := NodeBuildNew("node compiler", c)

			flags, err := b.generateEnvVariables()

			if !errCmp(err, tt.expected.err) {
				t.Errorf(cmp.Diff(err, tt.expected.err))
			}
			if err != nil {
				return
			}
			// Note: generated env variables contain the process's env variables too.
			expectedFlags := append(os.Environ(), tt.expected.flags...)
			sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if !cmp.Equal(flags, expectedFlags, sorted) {
				t.Errorf(cmp.Diff(flags, expectedFlags))
			}
		})
	}
}
*/

// TODO: discuss flags supported for Node.js builder.
/*
func Test_generateFlags(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		flags    []string
		expected error
	}{
		{
			name:     "valid flags",
			flags:    []string{"-race", "-x"},
			expected: nil,
		},
		{
			name:     "invalid -mod flags",
			flags:    []string{"-mod=whatever", "-x"},
			expected: errorUnsupportedArguments,
		},
		{
			name: "invalid random flags",
			flags: []string{
				"-a", "-race", "-msan", "-asan",
				"-v", "-x", "-buildinfo", "-buildmode",
				"-buildvcs", "-compiler", "-gccgoflags",
				"-gcflags", "-ldflags", "-linkshared",
				"-tags", "-trimpath", "bla",
			},
			expected: errorUnsupportedArguments,
		},
		{
			name: "valid all flags",
			flags: []string{
				"-a", "-race", "-msan", "-asan",
				"-v", "-x", "-buildinfo", "-buildmode",
				"-buildvcs", "-compiler", "-gccgoflags",
				"-gcflags", "-ldflags", "-linkshared",
				"-tags", "-trimpath",
			},
			expected: nil,
		},
	}

	for _, tt := range tests {
		tt := tt // Re-initializing variable so it is not changed while executing the closure below
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()

			cfg := goReleaserConfigFile{
				Version: 1,
				Flags:   tt.flags,
			}
			c, err := fromConfig(&cfg)
			if err != nil {
				t.Errorf("fromConfig: %v", err)
			}
			b := GoBuildNew("gocompiler", c)

			flags, err := b.generateFlags()
			expectedFlags := append([]string{"gocompiler", "build", "-mod=vendor"}, tt.flags...)
			fmt.Println(err)
			if !errCmp(err, tt.expected) {
				t.Errorf(cmp.Diff(err, tt.expected))
			}
			if err != nil {
				return
			}
			// Note: generated env variables contain the process's env variables too.
			sorted := cmpopts.SortSlices(func(a, b string) bool { return a < b })
			if !cmp.Equal(flags, expectedFlags, sorted) {
				t.Errorf(cmp.Diff(flags, expectedFlags))
			}
		})
	}
}
*/
