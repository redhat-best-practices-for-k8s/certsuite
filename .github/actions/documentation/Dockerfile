FROM registry.access.redhat.com/ubi9/python-39:1-192.1723128185@sha256:dbaa6cb3c0e9f9e017aa9f3ed184a450e8103f07017c8061fbc65453924cf649

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
