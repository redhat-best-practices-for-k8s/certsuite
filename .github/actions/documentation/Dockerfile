FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:8112ecca0a3c25a964de3999463404f4386a5e291655600d05e645c6b29a389f

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
