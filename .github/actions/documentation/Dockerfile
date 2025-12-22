FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:739f23a2ce5ac4ed5e4de6ec60ff583598fcfccb0d4ab591ff1b9c77477b3f7a

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
