// Copyright (C) 2018-2019 Miquel Sabaté Solà <mikisabate@gmail.com>
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
	"bytes"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func getPath(subpath string) string {
	const pkg = "github.com/mssola/openhub"
	return filepath.Join(os.Getenv("GOPATH"), "src", pkg, subpath)
}

func assertSlice(t *testing.T, one, two []string) {
	if len(one) != len(two) {
		t.Fatalf("Expecting %v listeners, %v given", len(one), len(two))
	}

	for _, v := range two {
		found := false
		for _, a := range one {
			if a == v {
				found = true
				break
			}
		}
		if !found {
			t.Fatalf("Element '%v' could not be found in %#v", v, one)
		}
	}
}

func findListener(t *testing.T, list []Listener, name string) Listener {
	for _, v := range list {
		if v.Name == name {
			return v
		}
	}
	t.Fatalf("Expecting '%v' listener could not be found", name)
	return list[0]
}

func assertListener(t *testing.T, one, two Listener) {
	assertString(t, one.Name, two.Name)
	assertString(t, one.Project, two.Project)
	assertString(t, one.Distribution, two.Distribution)
	assertString(t, one.Architecture, two.Architecture)
	assertString(t, one.Package, two.Package)
	assertString(t, one.Repository, two.Repository)
	assertSlice(t, one.Tags, two.Tags)
}

func testListeners(t *testing.T, got, expect []Listener) {
	if len(got) != len(expect) {
		t.Fatalf("Expecting %v listeners, %v given", len(expect), len(got))
	}

	for _, v := range expect {
		list := findListener(t, got, v.Name)
		assertListener(t, list, v)
	}
}

func TestParseConfiguration(t *testing.T) {
	// Setting up log.
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	// Actual call.
	cfg, err := ParseConfiguration(
		getPath("test/noarchnodist.yml"),
		Credentials{
			Server:   "https://api.opensuse.org",
			User:     "mssola",
			Password: "password",
			Token:    "token",
		},
		Options{SingleShot: true},
	)

	// Test listeners.
	if err != nil {
		fmt.Printf("Error: %#v\n", err.Error())
		t.Fatalf("Expecting no errors")
	}
	testListeners(t, cfg.Listeners, []Listener{
		{
			Name:         "portus-head",
			Project:      "Virtualization:containers:Portus",
			Distribution: "openSUSE_Leap_15.0",
			Architecture: "x86_64",
			Package:      "portus",
			Repository:   "opensuse/portus",
			Tags:         []string{"head"},
		},
		{
			Name:         "portus-2.3",
			Project:      "Virtualization:containers:Portus:2.3",
			Distribution: "openSUSE_Leap_42.3",
			Architecture: "x86_64",
			Package:      "portus",
			Repository:   "opensuse/portus",
			Tags:         []string{"2.3", "latest"},
		},
	})

	// Test log.
	logged := strings.SplitN(buf.String(), "\n", 2)
	if !strings.Contains(logged[0], "assuming openSUSE_Leap_15.0") {
		t.Fatalf("Wrong log")
	}
	if !strings.Contains(logged[1], "assuming x86_64") {
		t.Fatalf("Wrong log")
	}
}

func TestParseConfigurationUnknownConfig(t *testing.T) {
	_, err := ParseConfiguration(
		getPath("examples/unknown.yml"),
		Credentials{
			Server:   "https://api.opensuse.org",
			User:     "mssola",
			Password: "password",
			Token:    "token",
		},
		Options{SingleShot: true},
	)
	if err == nil {
		t.Fatalf("Expecting errors")
	}
	if !strings.Contains(err.Error(), "no such file or directory") {
		t.Fatalf("Wrong error")
	}
}

func TestParseConfigurationWrongYAML(t *testing.T) {
	_, err := ParseConfiguration(
		getPath("test/bad.yml"),
		Credentials{
			Server:   "https://api.opensuse.org",
			User:     "mssola",
			Password: "password",
			Token:    "token",
		},
		Options{SingleShot: true},
	)
	if err == nil {
		t.Fatalf("Expecting errors")
	}
	if !strings.Contains(err.Error(), "unmarshal errors") {
		t.Fatalf("Wrong error")
	}
}

func TestParseConfigurationNoProject(t *testing.T) {
	_, err := ParseConfiguration(
		getPath("test/noproject.yml"),
		Credentials{
			Server:   "https://api.opensuse.org",
			User:     "mssola",
			Password: "password",
			Token:    "token",
		},
		Options{SingleShot: true},
	)
	if err == nil {
		t.Fatalf("Expecting errors")
	}
	if !strings.Contains(err.Error(), "does not provide a project!") {
		t.Fatalf("Wrong error")
	}
}

func TestParseConfigurationNoPackage(t *testing.T) {
	_, err := ParseConfiguration(
		getPath("test/nopackage.yml"),
		Credentials{
			Server:   "https://api.opensuse.org",
			User:     "mssola",
			Password: "password",
			Token:    "token",
		},
		Options{SingleShot: true},
	)
	if err == nil {
		t.Fatalf("Expecting errors")
	}
	if !strings.Contains(err.Error(), "does not provide a package!") {
		t.Fatalf("Wrong error")
	}
}

func TestParseConfigurationNoRepository(t *testing.T) {
	_, err := ParseConfiguration(
		getPath("test/norepository.yml"),
		Credentials{
			Server:   "https://api.opensuse.org",
			User:     "mssola",
			Password: "password",
			Token:    "token",
		},
		Options{SingleShot: true},
	)
	if err == nil {
		t.Fatalf("Expecting errors")
	}
	if !strings.Contains(err.Error(), "does not provide a repository!") {
		t.Fatalf("Wrong error")
	}
}

func TestParseConfigurationNoTags(t *testing.T) {
	_, err := ParseConfiguration(
		getPath("test/notags.yml"),
		Credentials{
			Server:   "https://api.opensuse.org",
			User:     "mssola",
			Password: "password",
			Token:    "token",
		},
		Options{SingleShot: true},
	)
	if err == nil {
		t.Fatalf("Expecting errors")
	}
	if !strings.Contains(err.Error(), "does not provide tags!") {
		t.Fatalf("Wrong error")
	}
}
