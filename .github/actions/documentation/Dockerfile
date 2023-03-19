FROM registry.access.redhat.com/ubi8/ubi:8.7

RUN dnf install -y python39-3.9.13 && dnf clean all

# Pin versions in pip.
# hadolint ignore=DL3013
RUN pip3 install --no-cache-dir mkdocs mkdocs-material markdown-include

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
