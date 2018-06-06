FROM opensuse/amd64:42.3
MAINTAINER Miquel Sabaté Solà <mikisabate@gmail.com>

COPY . /go/src/github.com/mssola/openhub

ENV GOPATH /go
ENV PATH $GOPATH/bin:$PATH

RUN zypper ar -f -p 10 -g obs://devel:languages:go obs-dlg && \
	zypper -n --gpg-auto-import-keys ref && zypper -n up && \
    zypper -n in git 'go>=1.9' make && \
    cd /go/src/github.com/mssola/openhub; make install && \
    # Clean
    zypper -n rm git make kbd-legacy && \
    zypper clean -a && \
    rm -r /go/src

ENTRYPOINT ["openhub"]
