FROM registry.access.redhat.com/ubi9/python-39:1-182.1717085980@sha256:70d06e0da3fd00e8f834ebf8596a543c4a6e439d6e70cf720e31203d559d26ab

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
