FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:8b817c728990c39e8e9a40a5b9ec0fc1ed78656b8cf052b1c06a88373ad1bde8

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
