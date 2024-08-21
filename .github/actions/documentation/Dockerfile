FROM registry.access.redhat.com/ubi9/python-39:1-192.1724040313@sha256:ee9788013fc3cc9e90330600c26c11c9a6dc21030824daaf4a38b5c659096f38

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
