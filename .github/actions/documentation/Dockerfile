FROM registry.access.redhat.com/ubi9/python-312@sha256:bf2f5a3cf0c1542d2c90e729bef7e79df746e0331ab7fd19c86e3ff4792dc9f5

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
