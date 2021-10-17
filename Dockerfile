FROM fedora:latest
USER root
WORKDIR /app
COPY . ./
RUN dnf install -y go notmuch notmuch-devel python3 python3-pip && go build .
RUN pip3 install getmail6
RUN useradd -u 1026 -s /bin/sh imap
USER imap
ENTRYPOINT ["/app/go-imap-notmuch", "/config/config.yml"]
