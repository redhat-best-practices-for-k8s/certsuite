FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:32c5b684e23c4d1f51176770f3fa583a993f8c109086bfd3f74fcfbe59e2f871

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
