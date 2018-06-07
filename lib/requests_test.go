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
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"
	"time"
)

type testOptions struct {
	fail        bool
	timeout     bool
	decodeError bool
	tagsPushed  string
	n           int
}

func testOBS(opts *testOptions) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		opts.n = opts.n + 1

		if !strings.HasPrefix(r.URL.String(), "/build") {
			return
		}
		if r.Method != "GET" {
			return
		}
		if opts.fail {
			w.WriteHeader(401)
			return
		}
		if opts.timeout {
			time.Sleep(requestTimeout + (1 * time.Second))
		}

		// Base64 of user:password
		if r.Header["Authorization"][0] != "Basic dXNlcjpwYXNzd29yZA==" {
			return
		}
		if opts.decodeError {
			w.WriteHeader(200)
			fmt.Fprint(w, "<")
			return
		}

		if strings.HasSuffix(r.URL.String(), "/_status") {
			w.WriteHeader(200)
			fmt.Fprint(w, "<status package=\"portus\" code=\"succeeded\" />")
		} else if strings.HasSuffix(r.URL.String(), "/_buildinfo") {
			w.WriteHeader(200)
			fmt.Fprint(w, "<buildinfo><rev>1234</rev></buildinfo>")
		}
	}))
}

func getFromBody(req *http.Request) string {
	p := struct {
		Tag string `json:"docker_tag"`
	}{}

	if req.Body == nil {
		return ""
	}

	decoder := json.NewDecoder(req.Body)
	if err := decoder.Decode(&p); err != nil {
		return ""
	}
	return p.Tag
}

func testHub(opts *testOptions) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			return
		}
		if opts.fail {
			w.WriteHeader(401)
			return
		}
		if opts.timeout {
			time.Sleep(requestTimeout + (1 * time.Second))
		}

		str := getFromBody(r)
		opts.tagsPushed = opts.tagsPushed + "-" + str

		w.WriteHeader(200)
	}))
}

func TestStatusSucceededOK(t *testing.T) {
	server := testOBS(&testOptions{
		fail:        false,
		timeout:     false,
		decodeError: false,
	})
	defer server.Close()

	res := statusSucceeded(&Configuration{
		Server:   server.URL,
		User:     "user",
		Password: "password",
	}, Listener{})

	if !res {
		t.Fatalf("Expecting to be OK")
	}
}

func TestStatusSucceededTimeout(t *testing.T) {
	original := requestTimeout
	requestTimeout = 1 * time.Second
	defer func() { requestTimeout = original }()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	opts := &testOptions{
		fail:        false,
		timeout:     true,
		decodeError: false,
	}
	server := testOBS(opts)
	defer server.Close()
	dockerHub = server.URL + "/"

	res := statusSucceeded(&Configuration{
		Server:   server.URL,
		User:     "user",
		Password: "password",
	}, Listener{})
	if res {
		t.Fatalf("Expecting NOT to be OK")
	}

	logged := buf.String()
	if !strings.Contains(logged, "request canceled") {
		t.Fatalf("Wrong log")
	}
}

func TestStatusSucceededBadRequest(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	server := testOBS(&testOptions{
		fail:        true,
		timeout:     false,
		decodeError: false,
	})
	defer server.Close()

	res := statusSucceeded(&Configuration{
		Server:   server.URL,
		User:     "user",
		Password: "password",
	}, Listener{})

	if res {
		t.Fatalf("Expecting NOT to be OK")
	}

	logged := buf.String()
	if !strings.Contains(logged, "Status 401 when checking the status") {
		t.Fatalf("Wrong log")
	}
}

func TestStatusSucceededBadXML(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	server := testOBS(&testOptions{
		fail:        false,
		timeout:     false,
		decodeError: true,
	})
	defer server.Close()

	res := statusSucceeded(&Configuration{
		Server:   server.URL,
		User:     "user",
		Password: "password",
	}, Listener{})

	if res {
		t.Fatalf("Expecting NOT to be OK")
	}

	logged := buf.String()
	if !strings.Contains(logged, "XML syntax error") {
		t.Fatalf("Wrong log")
	}
}

func TestFetchRevisionOK(t *testing.T) {
	server := testOBS(&testOptions{
		fail:        false,
		timeout:     false,
		decodeError: false,
	})
	defer server.Close()

	res := fetchRevision(&Configuration{
		Server:   server.URL,
		User:     "user",
		Password: "password",
	}, Listener{})

	if res != "1234" {
		t.Fatalf("Expecting to be OK")
	}
}

func TestFetchRevisionBadRequest(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	server := testOBS(&testOptions{
		fail:        true,
		timeout:     false,
		decodeError: false,
	})
	defer server.Close()

	res := fetchRevision(&Configuration{
		Server:   server.URL,
		User:     "user",
		Password: "password",
	}, Listener{})

	if res != "" {
		t.Fatalf("Expecting NOT to be OK")
	}

	logged := buf.String()
	if !strings.Contains(logged, "Status 401 when checking the status") {
		t.Fatalf("Wrong log")
	}
}

func TestFetchRevisionBadXML(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	server := testOBS(&testOptions{
		fail:        false,
		timeout:     false,
		decodeError: true,
	})
	defer server.Close()

	res := fetchRevision(&Configuration{
		Server:   server.URL,
		User:     "user",
		Password: "password",
	}, Listener{})

	if res != "" {
		t.Fatalf("Expecting NOT to be OK")
	}

	logged := buf.String()
	if !strings.Contains(logged, "XML syntax error") {
		t.Fatalf("Wrong log")
	}
}

func TestHubOK(t *testing.T) {
	opts := &testOptions{
		fail:    false,
		timeout: false,
	}
	server := testHub(opts)
	defer server.Close()
	dockerHub = server.URL + "/"

	res := updateHub("1234", "example/repo", []string{"latest", "one"})
	if !res {
		t.Fatalf("Expecting to be OK")
	}
	if opts.tagsPushed != "-latest-one" {
		t.Fatalf("Not all tags were pushed")
	}
}

func TestHubTimeout(t *testing.T) {
	original := requestTimeout
	requestTimeout = 1 * time.Second
	defer func() { requestTimeout = original }()

	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	opts := &testOptions{
		fail:    false,
		timeout: true,
	}
	server := testHub(opts)
	defer server.Close()
	dockerHub = server.URL + "/"

	res := updateHub("1234", "example/repo", []string{"latest", "one"})
	if res {
		t.Fatalf("Expecting NOT to be OK")
	}

	logged := buf.String()
	if !strings.Contains(logged, "request canceled") {
		t.Fatalf("Wrong log")
	}
}

func TestHubBadRequest(t *testing.T) {
	var buf bytes.Buffer
	log.SetOutput(&buf)
	defer func() { log.SetOutput(os.Stderr) }()

	opts := &testOptions{
		fail:    true,
		timeout: false,
	}
	server := testHub(opts)
	defer server.Close()
	dockerHub = server.URL + "/"

	res := updateHub("1234", "example/repo", []string{"latest", "one"})
	if res {
		t.Fatalf("Expecting NOT to be OK")
	}

	logged := buf.String()
	if !strings.Contains(logged, "Status 401") {
		t.Fatalf("Wrong log")
	}
}
