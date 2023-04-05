FROM umputun/baseimage:buildgo-latest as build

ENV GOFLAGS="-mod=vendor"
ENV CGO_ENABLED=0

ADD . /build
WORKDIR /build

RUN cd build && go build -o /build/publisher -ldflags "-s -w"


FROM umputun/baseimage:app-latest

RUN apk add --no-cache openssh-client

COPY --from=build /build/publisher /srv/publisher
COPY build/ansible.yml /srv/ansible.yml
COPY build/cover.jpg /srv/cover.jpg

WORKDIR /srv

# Add the umputun user
RUN adduser -D -u 1000 umputun

# Set the owner of the files to the umputun user
RUN chown -R umputun:umputun /srv

# Add SSH known_hosts
RUN mkdir -p /home/umputun/.ssh && ssh-keyscan podcast.umputun.com >> /home/umputun/.ssh/known_hosts && \
    echo "StrictHostKeyChecking no" >> /home/umputun/.ssh/config

# Set the entrypoint to run the process under the umputun user
USER umputun

ENTRYPOINT ["/srv/publisher"]
