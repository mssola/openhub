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
	"bytes"
	"log"
	"os"
	"strings"
	"testing"
)

func assertString(t *testing.T, exp, got string) {
	if got != exp {
		t.Fatalf("Expecting '%v'; got '%v'", exp, got)
	}
}

func TestJoinTagsNone(t *testing.T) {
	assertString(t, "", joinTags([]string{}))
}

func TestJoinTagsSingle(t *testing.T) {
	assertString(t, "'tag'", joinTags([]string{"tag"}))
}

func TestJoinTagsMultiple(t *testing.T) {
	assertString(t, "'tag', 'tag1'", joinTags([]string{"tag", "tag1"}))
}

func TestSyncSingleShotOK(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	// Setting up servers.
	obs := testOBS(&testOptions{
		fail:        false,
		timeout:     false,
		decodeError: false,
	})
	defer obs.Close()

	opts := &testOptions{
		fail:    false,
		timeout: false,
	}
	hub := testHub(opts)
	defer hub.Close()
	dockerHub = hub.URL + "/"

	// Call Sync.
	res := Sync(&Configuration{
		Server:     obs.URL,
		User:       "user",
		Password:   "password",
		Token:      "token",
		SingleShot: true,
		Listeners: []Listener{
			{
				Name:         "portus-2.3",
				Project:      "Virtualization:containers:Portus:2.3",
				Distribution: "openSUSE_Leap_42.3",
				Architecture: "x86_64",
				Package:      "portus",
				Repository:   "opensuse/portus",
				Tags:         []string{"2.3", "latest"},
			},
		},
	})

	if res != nil {
		t.Fatalf("An error occurred: %#v\n", res.Error())
	}
	if opts.tagsPushed != "-2.3-latest" {
		t.Fatalf("Not all tags were pushed")
	}
	logged := buf.String()
	msg := "Updated to revision '1234' the tags: '2.3', 'latest'; for repository 'opensuse/portus'"
	if !strings.Contains(logged, msg) {
		t.Fatalf("Wrong log")
	}
}

func TestSyncOBSFails(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	// Setting up servers.
	obs := testOBS(&testOptions{
		fail:        true,
		timeout:     false,
		decodeError: false,
	})
	defer obs.Close()

	opts := &testOptions{
		fail:    false,
		timeout: false,
	}
	hub := testHub(opts)
	defer hub.Close()
	dockerHub = hub.URL + "/"

	// Call Sync.
	res := Sync(&Configuration{
		Server:     obs.URL,
		User:       "user",
		Password:   "password",
		Token:      "token",
		SingleShot: true,
		Listeners: []Listener{
			{
				Name:         "portus-2.3",
				Project:      "Virtualization:containers:Portus:2.3",
				Distribution: "openSUSE_Leap_42.3",
				Architecture: "x86_64",
				Package:      "portus",
				Repository:   "opensuse/portus",
				Tags:         []string{"2.3", "latest"},
			},
		},
	})

	if res != nil {
		t.Fatalf("An error occurred: %#v\n", res.Error())
	}
	if opts.tagsPushed != "" {
		t.Fatalf("Some tags were pushed")
	}
}

func TestSyncHubFails(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	// Setting up servers.
	obs := testOBS(&testOptions{
		fail:        false,
		timeout:     false,
		decodeError: false,
	})
	defer obs.Close()

	opts := &testOptions{
		fail:    true,
		timeout: false,
	}
	hub := testHub(opts)
	defer hub.Close()
	dockerHub = hub.URL + "/"

	// Call Sync.
	res := Sync(&Configuration{
		Server:     obs.URL,
		User:       "user",
		Password:   "password",
		Token:      "token",
		SingleShot: true,
		Listeners: []Listener{
			{
				Name:         "portus-2.3",
				Project:      "Virtualization:containers:Portus:2.3",
				Distribution: "openSUSE_Leap_42.3",
				Architecture: "x86_64",
				Package:      "portus",
				Repository:   "opensuse/portus",
				Tags:         []string{"2.3", "latest"},
			},
		},
	})

	if res != nil {
		t.Fatalf("An error occurred: %#v\n", res.Error())
	}
	if opts.tagsPushed != "" {
		t.Fatalf("Some tags were pushed")
	}
}
