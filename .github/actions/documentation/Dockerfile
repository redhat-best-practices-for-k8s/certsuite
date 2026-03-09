FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:4b8183c9ec796f3f4b4e0a9998ab21a89af9da8ea1b476dc3a0c8473c8fac0dd

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
