FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:d3a8113accb685a6b4c6fb23fc465821a408869e877dbba3e07283f3c3cdbf4e

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
