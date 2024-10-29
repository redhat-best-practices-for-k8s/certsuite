FROM registry.access.redhat.com/ubi9/python-39:1-197.1729767844@sha256:e7f5b60728d8e71588272e28b88c2e8fe9e63609347fb379b8c1625df886e189

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
