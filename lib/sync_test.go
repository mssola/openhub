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

import "testing"

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
