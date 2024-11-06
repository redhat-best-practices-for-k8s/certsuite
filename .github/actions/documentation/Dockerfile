FROM registry.access.redhat.com/ubi9/python-39:1-197.1729776550@sha256:db5a0fc74728e23de67f38fa864c9038ce61a93066599c1cd250e41a20507ce8

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
