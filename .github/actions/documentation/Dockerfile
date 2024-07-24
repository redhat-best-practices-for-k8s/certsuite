FROM registry.access.redhat.com/ubi9/python-39:1-192@sha256:01c9e53b32acd96f9fe7781727140df6868c91ebc916ed95dc58999fbf4d8ddd

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
