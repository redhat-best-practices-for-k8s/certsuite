FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:fecb260bdb384c3ae58925b6b6b3a8b519fd53ad88ca98e3bc44fa4bbd4722ea

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
