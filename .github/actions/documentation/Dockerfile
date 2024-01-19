FROM registry.access.redhat.com/ubi9/ubi:9.3-1476

# hadolint ignore=DL3041
RUN dnf install -y python39 && dnf clean all

# Pin versions in pip.
# hadolint ignore=DL3013
RUN pip3 install --no-cache-dir mkdocs mkdocs-material markdown-include

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
