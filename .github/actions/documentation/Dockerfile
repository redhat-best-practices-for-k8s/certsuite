FROM registry.access.redhat.com/ubi9/python-39:1-182.1714671018@sha256:f33083d843898a6ffae35d91d6ff4bf6cfe0d9585c7897b04834cda30ede5b5c

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
