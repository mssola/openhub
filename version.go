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

package main

import "fmt"

var gitCommit, version string

func versionString() string {
	str := version

	if gitCommit != "" {
		str += fmt.Sprintf(" with commit '%v'", gitCommit)
	}
	return fmt.Sprintf(`%v.
Copyright (C) 2018-2019 Miquel Sabaté Solà <mikisabate@gmail.com>
License GPLv3+: GNU GPL version 3 or later "<http://gnu.org/licenses/gpl.html>.
This is free software: you are free to change and redistribute it.
There is NO WARRANTY, to the extent permitted by law.`, str)
}
