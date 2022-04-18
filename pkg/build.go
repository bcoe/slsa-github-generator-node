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
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"syscall"
)

var (
	errorEnvVariableNameEmpty      = errors.New("env variable empty or not set")
	errorUnsupportedArguments      = errors.New("argument not supported")
	errorInvalidEnvArgument        = errors.New("invalid env passed via argument")
	errorEnvVariableNameNotAllowed = errors.New("env variable not allowed")
	errorInvalidFilename           = errors.New("invalid filename")
	errorEmptyFilename             = errors.New("filename is not set")
)

// See `npm pack --help`.

var allowedBuildArgs = map[string]bool{
	"--workspace":              true,
	"--include-workspace-root": true,
}

var allowedEnvVariablePrefix = map[string]bool{
	"NODE_": true,
}

type NodeBuild struct {
	pkgJson *PkgJsonConfig
	node   string
	npm    string
	// Note: static env variables are contained in cfg.Env.
	argEnv map[string]string
}

func NodeBuildNew(node string, npm string, pkgJson *PkgJsonConfig) *NodeBuild {
	c := NodeBuild{
		pkgJson: pkgJson,
		node:   node,
		npm:    npm,
		argEnv:  make(map[string]string),
	}

	return &c
}

func (b *NodeBuild) Run(dry bool) error {
	// Set flags.
	flags, err := b.generateFlags()
	if err != nil {
		return err
	}

	// Generate env variables.
	envs, err := b.generateEnvVariables()
	if err != nil {
		return err
	}

	com := append([]string{b.node, b.npm, "pack"}, flags...)

	// A dry run prints the information that is trusted, before
	// the compiler is invoked.
	if dry {
		// Generate filename.
		filename, err := b.generateOutputFilename()
		if err != nil {
			return err
		}

		// Share the resolved name of the binary.
		fmt.Printf("::set-output name=node-package-name::%s\n", filename)
		command, err := marshallList(com)
		if err != nil {
			return err
		}
		// Share the command used.
		fmt.Printf("::set-output name=node-command::%s\n", command)

		env, err := b.generateCommandEnvVariables()
		if err != nil {
			return err
		}

		menv, err := marshallList(env)
		if err != nil {
			return err
		}

		// Share the env variables used.
		fmt.Printf("::set-output name=node-env::%s\n", menv)
		return nil
	}

	fmt.Println("command", com)
	fmt.Println("env", envs)
	return syscall.Exec(b.node, com, envs)
}

func marshallList(args []string) (string, error) {
	jsonData, err := json.Marshal(args)
	if err != nil {
		return "", fmt.Errorf("json.Marshal: %w", err)
	}

	encoded := base64.StdEncoding.EncodeToString(jsonData)
	if err != nil {
		return "", fmt.Errorf("base64.StdEncoding.DecodeString: %w", err)
	}
	return encoded, nil
}

func (b *NodeBuild) generateCommandEnvVariables() ([]string, error) {
	var env []string

	// TODO: what's the equivalent of Goos for Node.js?
	// if b.cfg.Goos == "" {
	//	return nil, fmt.Errorf("%w: %s", errorEnvVariableNameEmpty, "GOOS")
	//}
	//env = append(env, fmt.Sprintf("GOOS=%s", b.cfg.Goos))

	// TODO: what environment variables might we allow list for Node.js.
	// Set env variables from config file.
	// for k, v := range b.cfg.Env {
	//	if !isAllowedEnvVariable(k) {
	//		return env, fmt.Errorf("%w: %s", errorEnvVariableNameNotAllowed, v)
	//	}

	//	env = append(env, fmt.Sprintf("%s=%s", k, v))
	//}

	return env, nil
}

func (b *NodeBuild) generateEnvVariables() ([]string, error) {
	env := os.Environ()

	cenv, err := b.generateCommandEnvVariables()
	if err != nil {
		return cenv, err
	}

	env = append(env, cenv...)

	return env, nil
}

func (b *NodeBuild) SetArgEnvVariables(envs string) error {
	// Notes:
	// - I've tried running the re-usable workflow in a step
	// and set the env variable in a previous step, but found that a re-usable workflow is not
	// allowed to run in a step; they have to run as `job.uses`. Using `job.env` with `job.uses`
	// is not allowed.
	// - We don't want to allow env variables set in the workflow because of injections
	// e.g. LD_PRELOAD, etc.
	if envs == "" {
		return nil
	}

	for _, e := range strings.Split(envs, ",") {
		s := strings.Trim(e, " ")

		sp := strings.Split(s, ":")
		if len(sp) != 2 {
			return fmt.Errorf("%w: %s", errorInvalidEnvArgument, s)
		}
		name := strings.Trim(sp[0], " ")
		value := strings.Trim(sp[1], " ")

		fmt.Printf("arg env: %s:%s\n", name, value)
		b.argEnv[name] = value

	}
	return nil
}

func (b *NodeBuild) generateOutputFilename() (string, error) {
	// TODO: validate that "name", "version", are not nil.

	return b.pkgJson.Name + "-" + b.pkgJson.Version + ".tgz", nil
}

func (b *NodeBuild) generateFlags() ([]string, error) {
	// -x
	flags := []string{}

	// TODO: add support for flags in Node.js build.
	//for _, v := range b.cfg.Flags {
	//	if !isAllowedArg(v) {
	//		return nil, fmt.Errorf("%w: %s", errorUnsupportedArguments, v)
	//	}
	//	flags = append(flags, v)
	//}
	return flags, nil
}

func isAllowedArg(arg string) bool {
	for k := range allowedBuildArgs {
		if strings.HasPrefix(arg, k) {
			return true
		}
	}
	return false
}

// Check if the env variable is allowed. We want to avoid
// variable injection, e.g. LD_PRELOAD, etc.
// See an overview in https://www.hale-legacy.com/class/security/s20/handout/slides-env-vars.pdf.
func isAllowedEnvVariable(name string) bool {
	for k := range allowedEnvVariablePrefix {
		if strings.HasPrefix(name, k) {
			return true
		}
	}
	return false
}
