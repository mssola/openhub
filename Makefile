# Copyright (C) 2018 Miquel Sabaté Solà <mikisabate@gmail.com>
#
# This program is free software: you can redistribute it and/or modify
# it under the terms of the GNU General Public License as published by
# the Free Software Foundation, either version 3 of the License, or
# (at your option) any later version.
#
# This program is distributed in the hope that it will be useful,
# but WITHOUT ANY WARRANTY; without even the implied warranty of
# MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
# GNU General Public License for more details.
#
# You should have received a copy of the GNU General Public License
# along with this program.  If not, see <http://www.gnu.org/licenses/>.

# This Makefile has taken lots of ideas and code from openSUSE/umoci by Aleksa Sarai.

# Use bash, so that we can do process substitution.
SHELL = /bin/bash

GO ?= go
CMD ?= openhub
GO_SRC = $(shell find . -name \*.go)

# Version information.
GIT := $(shell command -v git 2> /dev/null)
VERSION := $(shell cat ./VERSION)
COMMIT_NO := $(if $GIT,$(shell git rev-parse HEAD 2> /dev/null),"")
ifdef $GIT
  COMMIT := $(if $(shell git status --porcelain --untracked-files=no),"${COMMIT_NO}-dirty","${COMMIT_NO}")
else
  COMMIT := "${COMMIT_NO}"
endif

# Build flags and settings.
BUILD_FLAGS ?=
DYN_BUILD_FLAGS := $(BUILD_FLAGS) -buildmode=pie -ldflags "-s -w -X main.gitCommit=${COMMIT} -X main.version=${VERSION}" -tags "$(BUILDTAGS)"

.DEFAULT: openhub
openhub: $(GO_SRC)
	@$(GO) build ${DYN_BUILD_FLAGS} -o $(CMD)

.PHONY: install
install: $(GO_SRC)
	@$(GO) install -v ${DYN_BUILD_FLAGS} .

.PHONY: clean
clean: clean-binary

clean-binary:
	@rm -rf $(CMD)

#
# Unit tests.
#

.PHONY: test
test: clean
	$(GO) test -v ./...

#
# Validation tools.
#

.PHONY: validate-go
validate-go:
	@which gofmt >/dev/null 2>/dev/null || (echo "ERROR: gofmt not found." && false)
	test -z "$$(gofmt -s -l . | grep -vE '^vendor/' | tee /dev/stderr)"
	@which golint >/dev/null 2>/dev/null || (echo "ERROR: golint not found." && false)
	test -z "$$(golint . | grep -v vendor | tee /dev/stderr)"
	@go doc cmd/vet >/dev/null 2>/dev/null || (echo "ERROR: go vet not found." && false)
	test -z "$$(go vet . | grep -v vendor | tee /dev/stderr)"

.PHONY: validate
validate: validate-go

#
# Docker image
#

.PHONY: image
image:
	docker build -t mssola/$(CMD):$(VERSION) .

#
# Travis-CI
#

.PHONY: ci
ci: openhub validate test
