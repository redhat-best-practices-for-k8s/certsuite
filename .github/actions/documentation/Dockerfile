FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:351f14d057c0c017b9153c81ef9a7a1a8c1356a4c63ec92e9db5f6397c027a89

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
