FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:bdc09052ddc6fca1b707b12672ce97c4ac538f412841a4d90c82e50d1bbf154f

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
