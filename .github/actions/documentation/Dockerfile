FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:c3d30a724ff5781cd470979300aa9b2700098c3a8ca1b0e4e5c75080294f33e5

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
