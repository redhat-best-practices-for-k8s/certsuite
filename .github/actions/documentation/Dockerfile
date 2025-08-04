FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:d472177f6de3e683f6b68fb0a45f0c0644302e89888933f85520b95c16ef7cfe

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
