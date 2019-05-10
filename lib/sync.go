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
	"log"
	"strings"
	"sync"
	"time"
)

type state struct {
	done      bool
	revisions map[string]string
}

var syncTimeout = 5 * time.Minute

func Sync(cfg *Configuration) error {
	st := &state{
		done:      false,
		revisions: make(map[string]string),
	}

	performSync(cfg, st)
	if cfg.SingleShot {
		log.Printf("Only one execution was needed, stopping...")
		return nil
	}

	log.Printf("Listening...")
	for {
		select {
		case <-time.After(syncTimeout):
			if !st.done {
				log.Printf("Previous execution is not done, waiting...")
			} else {
				st.done = false
				performSync(cfg, st)
			}
		}
	}
	return nil
}

func performSync(cfg *Configuration, st *state) {
	var waitGroup sync.WaitGroup
	waitGroup.Add(len(cfg.Listeners))
	defer func() {
		waitGroup.Wait()
		st.done = true
	}()

	for _, v := range cfg.Listeners {
		go func(list Listener) {
			defer waitGroup.Done()
			synchronize(cfg, list, st)
		}(v)
	}
}

func synchronize(cfg *Configuration, list Listener, st *state) {
	if !statusSucceeded(cfg, list) {
		return
	}
	rev := fetchRevision(cfg, list)
	if rev == "" {
		return
	}

	val, ok := st.revisions[list.Name]
	if !ok || val != rev {
		if list.LocalBuild {
			if err := buildAndPush(cfg, list); err != nil {
				log.Printf("Failed to update to revision '%v' for the tags: %v; for repository '%v': %v",
					rev, joinTags(list.Tags), list.Repository, err)
			}
		} else if updateHub(cfg.Token, list.Repository, list.Tags) {
			log.Printf("Updated to revision '%v' the tags: %v; for repository '%v'",
				rev, joinTags(list.Tags), list.Repository)
		}
		st.revisions[list.Name] = rev
	} else {
		log.Printf("%v: everything up-to-date, skipping...", list.Name)
	}
}

func joinTags(tags []string) string {
	res := []string{}
	for _, v := range tags {
		res = append(res, "'"+v+"'")
	}
	return strings.Join(res, ", ")
}
