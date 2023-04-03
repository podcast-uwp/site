FROM umputun/baseimage:buildgo-latest as build

ENV GOFLAGS="-mod=vendor"
ENV CGO_ENABLED=0

ADD . /build
WORKDIR /build

RUN go build -o /build/publisher -ldflags "-s -w"
RUN GOOS=darwin GOARCH=arm64 go build -o /build/publisher.osx -ldflags "-s -w"

FROM umputun/baseimage:app-latest

RUN apk add --no-cache openssh-client ansible && rm -rf /var/cache/apk/*

COPY --from=build /build/publisher /srv/publisher
COPY --from=build /build/publisher.osx /srv/publisher.osx
COPY ansible.yml /srv/ansible.yml
COPY cover.jpg /srv/cover.jpg

WORKDIR /srv

USER app
RUN mkdir -p /home/app/.ssh && ssh-keyscan podcast.umputun.com >> /home/app/.ssh/known_hosts && \
    echo "StrictHostKeyChecking no" >> /home/app/.ssh/config

