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
	"encoding/xml"
	"io/ioutil"
	"log"
	"net/http"
	"path/filepath"
	"time"
)

const requestTimeout = 15 * time.Second

func request(cfg *Configuration, method, endpoint string) (*http.Response, error) {
	client := http.Client{Timeout: requestTimeout}

	req, _ := http.NewRequest(method, cfg.Server+endpoint, nil)
	req.SetBasicAuth(cfg.User, cfg.Password)

	return client.Do(req)
}

func safeRequest(pre, post string, cfg *Configuration, list Listener) (*http.Response, bool) {
	endpoint := filepath.Join(pre, list.Project, list.Distribution,
		list.Architecture, list.Package, post)
	resp, err := request(cfg, "GET", endpoint)
	if err != nil {
		log.Printf("error: %v", err)
		return nil, false
	}

	if resp.StatusCode != http.StatusOK {
		log.Printf("Status %v when checking the status", resp.StatusCode)
		return nil, false
	}
	return resp, true
}

func statusSucceeded(cfg *Configuration, list Listener) bool {
	resp, b := safeRequest("/build", "_status", cfg, list)
	if !b {
		return b
	}

	status := struct {
		XMLName xml.Name `xml:"status"`
		Package string   `xml:"package,attr"`
		Code    string   `xml:"code,attr"`
	}{}

	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&status); err != nil {
		log.Printf("error: %v", err)
		return false
	}
	return status.Code == "succeeded"
}

func fetchRevision(cfg *Configuration, list Listener) string {
	resp, b := safeRequest("/build", "_buildinfo", cfg, list)
	if !b {
		return ""
	}

	info := struct {
		XMLName  xml.Name `xml:"buildinfo"`
		Revision string   `xml:"rev"`
	}{}

	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&info); err != nil {
		log.Printf("error: %v", err)
		return ""
	}
	return info.Revision
}

func updateHub(token, repository string, tags []string) bool {
	client := http.Client{Timeout: requestTimeout}
	url := "https://registry.hub.docker.com/u/" + repository + "/trigger/" + token + "/"

	for _, tag := range tags {
		reader := bytes.NewBuffer([]byte("{\"docker_tag\": \"" + tag + "\"}"))
		req, _ := http.NewRequest("POST", url, reader)
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			log.Printf("error: %v", err)
			return false
		}
		if resp.StatusCode != http.StatusOK {
			log.Printf("Status %v when updating tag '%v' on Docker Hub", resp.StatusCode, tag)
			b, _ := ioutil.ReadAll(resp.Body)
			log.Printf("Given response: %v", string(b))
			return false
		}
	}
	return true
}
