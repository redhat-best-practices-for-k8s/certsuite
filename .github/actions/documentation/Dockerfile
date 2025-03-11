FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:6f89c966a1939d3fcd8919f1e823f1794721e68fb3b31388230529ff622eebef

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
