FROM fedora:latest
WORKDIR /app
COPY . ./
RUN dnf install -y go notmuch notmuch-devel python3 python3-pip && go build .
RUN pip3 install getmail6
ENTRYPOINT ["/app/go-imap-notmuch", "/config/config.yml"]
