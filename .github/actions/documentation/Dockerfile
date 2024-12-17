FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:daa95bd4459da0314c06a918ead906be049f74f2f19850cf259f06761ddb3979

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
