FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:e9cefd15e474e7782e6f5c9a1563c98d61aa5fd14975b498aefdeba583178d7d

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
