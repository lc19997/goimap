FROM fedora:latest
WORKDIR /app
COPY . ./
RUN dnf install -y go notmuch notmuch-devel && go build .
ENTRYPOINT ["/app/go-imap-notmuch /config/config.yml"]
