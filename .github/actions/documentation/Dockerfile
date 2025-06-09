FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:31ffcf9c19b329ef8753472a7095106c883baf4d26522d87fbbf011e6152d429

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
