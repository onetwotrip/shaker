ARG GO_VERSION=1.18
ARG ALPINE_VERSION=3.17

FROM golang:${GO_VERSION}-alpine${ALPINE_VERSION}
ARG USERNAME=vscode
ARG USER_UID=1000
ARG USER_GID=1000

# Setup user
RUN adduser $USERNAME -s /bin/sh -D -u $USER_UID $USER_GID && \
    mkdir -p /etc/sudoers.d && \
    echo $USERNAME ALL=\(root\) NOPASSWD:ALL > /etc/sudoers.d/$USERNAME && \
    chmod 0440 /etc/sudoers.d/$USERNAME

# Install packages and Go language server
RUN apk add -q --update --progress --no-cache git sudo openssh-client build-base curl vim
RUN go clean -modcache

USER $USERNAME
ENV LANG=en_US.UTF-8

ENV GOPATH=/go
ENV GO111MODULE=on

RUN go install -v github.com/go-delve/delve/cmd/dlv@latest
RUN go install -v golang.org/x/tools/gopls@latest

ENV CONFIG="/home/vscode/workspaces/shaker/dev/config.yml"