FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:c934d87840b237cb56f6ff87cbc963cc64e7156337edd3040452f1be817f2891

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
