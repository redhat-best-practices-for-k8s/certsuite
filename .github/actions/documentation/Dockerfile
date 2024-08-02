FROM registry.access.redhat.com/ubi9/python-39:1-192.1722518946@sha256:0176b477075984d5a502253f951d2502f0763c551275f9585ac515b9f241d73d

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
