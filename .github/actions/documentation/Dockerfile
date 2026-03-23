FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:6858f39fd97263227a10d0b79f098c22204b2f94a6062de465c0a31cc4040c0f

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
