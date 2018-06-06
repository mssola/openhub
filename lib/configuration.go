// Copyright (C) 2018 Miquel Sabaté Solà <mikisabate@gmail.com>
//
// This program is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// This program is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <http://www.gnu.org/licenses/>.

package lib

import (
	"fmt"
	"io/ioutil"
	"log"
	"path/filepath"

	"gopkg.in/urfave/cli.v1"
	"gopkg.in/yaml.v2"
)

const (
	defaultDistribution = "openSUSE_Leap_15.0"
	defaultArchitecutre = "x86_64"
)

// Configuration holds all the data relevant for this application to perform
// properly.
type Configuration struct {
	Server     string
	User       string
	Password   string
	Token      string
	SingleShot bool
	Listeners  []Listener
}

// Listener holds all the data relevant for services. That is, the OBS data and
// the Docker tags that relate to it.
type Listener struct {
	Name         string
	Project      string   `yaml:"project"`
	Distribution string   `yaml:"distribution"`
	Architecture string   `yaml:"architecture"`
	Package      string   `yaml:"package"`
	Repository   string   `yaml:"repository"`
	Tags         []string `yaml:"tags"`
}

// ConfigFile is the struct to be used when parsing the configuration.
type ConfigFile struct {
	Services map[string]Listener `yaml:"services,omitempty"`
}

// ParseConfiguration returns a proper Configuration object by taking into
// account the given flags and the configuration file.
func ParseConfiguration(ctx *cli.Context) (*Configuration, error) {
	listeners, err := parseConfiguration(ctx.Args().First())
	if err != nil {
		return nil, err
	}

	return &Configuration{
		Server:     ctx.String("server"),
		User:       ctx.String("user"),
		Password:   ctx.String("password"),
		Token:      ctx.String("token"),
		SingleShot: ctx.Bool("single-shot"),
		Listeners:  listeners,
	}, nil
}

// parseConfiguration returns a list of listeners by taking into account the
// configuration file.
func parseConfiguration(configurationPath string) ([]Listener, error) {
	data, err := readConfigFile(configurationPath)
	if err != nil {
		return nil, err
	}

	settings := ConfigFile{}
	if err = yaml.Unmarshal([]byte(data), &settings); err != nil {
		return nil, err
	}
	return sanitizeListeners(settings)
}

// readConfigFile returns the contents from the configuration file.
func readConfigFile(configurationPath string) ([]byte, error) {
	path, err := filepath.Abs(configurationPath)
	if err != nil {
		return nil, err
	}
	return ioutil.ReadFile(path)
}

// sanitizeListeners iterates over the parsed services and sanitizes their
// contents.
func sanitizeListeners(settings ConfigFile) ([]Listener, error) {
	listeners := []Listener{}

	for name, list := range settings.Services {
		if list.Project == "" {
			return nil, fmt.Errorf("%v service does not provide a project!", name)
		}
		if list.Package == "" {
			return nil, fmt.Errorf("%v service does not provide a package!", name)
		}
		if list.Repository == "" {
			return nil, fmt.Errorf("%v service does not provide a repository!", name)
		}
		if len(list.Tags) == 0 {
			return nil, fmt.Errorf("%v service does not provide tags!", name)
		}
		if list.Distribution == "" {
			list.Distribution = defaultDistribution
			log.Printf("%v service does not provide a distribution, assuming %v",
				name, defaultDistribution)
		}
		if list.Architecture == "" {
			list.Architecture = defaultArchitecutre
			log.Printf("%v service does not provide an architecture, assuming %v",
				name, defaultArchitecutre)
		}
		list.Name = name

		listeners = append(listeners, list)
	}

	return listeners, nil
}
