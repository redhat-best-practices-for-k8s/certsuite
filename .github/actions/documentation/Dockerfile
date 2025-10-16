FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:dee7605bfc0cc89b83075937cd48ff7b61bf5e396f8bd2ae13e7f7299be53326

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
