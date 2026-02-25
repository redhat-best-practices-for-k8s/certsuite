FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:f54df76a8d799877f5c95ac4dacb818137534d6d69e38556f08e8bcb417fc083

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
