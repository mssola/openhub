// Copyright (C) 2019 Miquel Sabaté Solà <mikisabate@gmail.com>
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
	"context"
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/client"
)

func buildAndPush(cfg *Configuration, list Listener) error {
	results, err := fetchPackage(cfg, list)
	if err != nil {
		return err
	}
	if err := loadImage(results, cfg, list); err != nil {
		return err
	}
	if err := pushImage(cfg, list); err != nil {
		return err
	}
	return nil
}

type Binary struct {
	Filename string `xml:"filename,attr"`
}

type Binaries struct {
	XMLName xml.Name `xml:"binarylist"`
	Results []Binary `xml:"binary"`
}

var (
	tarball = regexp.MustCompile(`\.docker\.tar$`)
	tarSha  = regexp.MustCompile(`\.docker\.tar\.sha256$`)
)

const (
	downloadDir = ".oh-downloads"
)

func downloadPackage(cfg *Configuration, list Listener, filename string) (string, error) {
	wd, _ := os.Getwd()
	downloadDirectory := filepath.Join(wd, downloadDir)
	_ = os.MkdirAll(downloadDirectory, 0744)

	log.Printf("Downloading file '%v' into '%v'", filename, downloadDirectory)

	resp, _ := safeRequest("/build", filename, cfg, list)
	defer resp.Body.Close()

	// Create the file
	dst := filepath.Join(downloadDirectory, filename)
	out, err := os.Create(dst)
	if err != nil {
		return "", err
	}
	defer out.Close()

	// Write the body to file
	_, err = io.Copy(out, resp.Body)
	return dst, err
}

func fetchPackage(cfg *Configuration, list Listener) ([]string, error) {
	packages := []string{}

	resp, b := safeRequest("/build", "", cfg, list)
	if !b {
		return packages, fmt.Errorf("could not fetch the binaries for the '%v' package", list.Package)
	}

	info := Binaries{}
	decoder := xml.NewDecoder(resp.Body)
	if err := decoder.Decode(&info); err != nil {
		// TODO
		return packages, err
	}

	for _, v := range info.Results {
		// TODO: guarantee that 0: tarball; 1: sha256 file
		if tarball.MatchString(v.Filename) || tarSha.MatchString(v.Filename) {
			if fp, err := downloadPackage(cfg, list, v.Filename); err != nil {
				return packages, fmt.Errorf("could not download '%v': %v", v.Filename, err)
			} else {
				packages = append(packages, fp)
			}
		}
	}
	return packages, nil
}

func removeImages(cli *client.Client, repo string, tags []string) {
	for _, v := range tags {
		img := repo + ":" + v
		cli.ImageRemove(context.Background(), img, types.ImageRemoveOptions{
			Force:         true,
			PruneChildren: true,
		})
	}
}

func loadImage(results []string, cfg *Configuration, list Listener) error {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return err
	}

	if !validImage(results) {
		return fmt.Errorf("'%v' does not match sha256 file '%v'", results[0], results[1])
	}

	removeImages(cli, list.Repository, list.Tags)

	fl, err := os.Open(results[0])
	if err != nil {
		return err
	}
	defer fl.Close()
	resp, err := cli.ImageLoad(context.Background(), fl, true)
	if err != nil {
		return err
	}

	defer resp.Body.Close()
	if resp.JSON {
		s := struct {
			Stream string `json:"stream"`
		}{}
		_ = json.NewDecoder(resp.Body).Decode(&s)
		log.Printf(strings.TrimSpace(s.Stream))
	}
	return nil
}

func pushImage(cfg *Configuration, list Listener) error {
	return nil
}

// TODO
func validImage(results []string) bool {
	return true
}
