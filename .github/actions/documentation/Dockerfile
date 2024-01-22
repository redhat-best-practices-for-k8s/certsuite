FROM registry.access.redhat.com/ubi9/python-39:1-161

# Pin versions in pip.
# hadolint ignore=DL3013
RUN pip install --no-cache-dir mkdocs mkdocs-material markdown-include

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
