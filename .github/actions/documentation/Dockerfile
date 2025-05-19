FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:f29fbe7a5990f195a89dd1d0ae8cfdb5fb57bbadfe950d5e6f541f5c2aaf8cb5

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
