FROM registry.access.redhat.com/ubi9/python-39:1-161@sha256:28cdbc26c697050370095ff2cffa1b8403f5501720b99f3796efeee896e1727d

# Pin versions in pip.
# hadolint ignore=DL3013
RUN pip install --no-cache-dir mkdocs mkdocs-material markdown-include

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
