FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:7d7dbfe4e208b4c71db010f2115d66aa6ba03034abe997ba274503f5283beb39

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
