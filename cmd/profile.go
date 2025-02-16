/*
 * Copyright © 2025 pegagio
 *
 * Permission is hereby granted, free of charge, to any person obtaining a copy
 * of this software and associated documentation files (the "Software"), to deal
 * in the Software without restriction, including without limitation the rights
 * to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
 * copies of the Software, and to permit persons to whom the Software is
 * furnished to do so, subject to the following conditions:
 *
 * The above copyright notice and this permission notice shall be included in
 * all copies or substantial portions of the Software.
 *
 * THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
 * IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
 * FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
 * AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
 * LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
 * OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
 * THE SOFTWARE.
 */

package cmd

import (
	"fmt"
	"os"
	"strings"

	"gopkg.in/yaml.v3"
)

type Profile struct {
	Variables  map[string]string `yaml:"variables"`
	Path       []string          `yaml:"path"`
	Aliases    map[string]string `yaml:"aliases"`
	Functions  map[string]string `yaml:"functions"`
	PreLoad    string            `yaml:"preload"`
	PostLoad   string            `yaml:"postload"`
	PreUnload  string            `yaml:"preunload"`
	PostUnload string            `yaml:"postunload"`
}

func ReadProfile(profilePath string) (Profile, error) {
	data, err := os.ReadFile(profilePath)
	if err != nil {
		return Profile{}, err
	}

	var config Profile
	err = yaml.Unmarshal(data, &config)
	return config, err
}

func (p *Profile) GenerateLoadScript() string {
	var script []string

	if p.PreLoad != "" {
		script = append(script, fmt.Sprintf("preload() {\n%s\n}\npreload", p.PreLoad))
	}

	for key, value := range p.Variables {
		script = append(script, fmt.Sprintf("export %s=\"%s\"", key, value))
	}

	if len(p.Path) > 0 {
		script = append(script, fmt.Sprintf("export PATH=\"%s:$PATH\"", strings.Join(p.Path, ":")))
	}

	for funcName, funcBody := range p.Functions {
		script = append(script, fmt.Sprintf("%s() {\n%s\n}", funcName, funcBody))
	}

	for aliasName, aliasCommand := range p.Aliases {
		script = append(script, fmt.Sprintf("alias %s='%s'", aliasName, aliasCommand))
	}

	if p.PostLoad != "" {
		script = append(script, fmt.Sprintf("postload() {\n%s\n}\npostload", p.PostLoad))
	}

	return strings.Join(script, "\n")
}

func (p *Profile) GenerateUnloadScript() string {
	var script []string

	if p.PreUnload != "" {
		script = append(script, fmt.Sprintf("preunload() {\n%s\n}\npreunload", p.PostUnload))
	}

	for key := range p.Variables {
		script = append(script, fmt.Sprintf("unset %s", key))
	}

	pathEnv := os.Getenv("PATH")
	for _, path := range p.Path {
		resolved := os.ExpandEnv(path)
		pathEnv = removeFromPath(pathEnv, resolved)
	}
	script = append(script, fmt.Sprintf("export PATH=%s", pathEnv))

	for functionName := range p.Functions {
		script = append(script, fmt.Sprintf("unset -f %s", functionName))
	}

	for aliasName := range p.Aliases {
		script = append(script, fmt.Sprintf("unalias %s 2>/dev/null || true", aliasName))
	}

	if p.PostUnload != "" {
		script = append(script, fmt.Sprintf("postunload() {\n%s\n}\npostunload", p.PostUnload))
	}

	return strings.Join(script, "\n")
}

func removeFromPath(path, profilePath string) string {
	// Split PATH into slices based on ":"
	pathSegments := strings.Split(path, ":")

	// Filter out segments that match profilePath
	var newPathSegments []string
	for _, segment := range pathSegments {
		if segment != profilePath {
			newPathSegments = append(newPathSegments, segment)
		}
	}

	// Join filtered segments back into a PATH string
	newPath := strings.Join(newPathSegments, ":")

	return newPath
}
