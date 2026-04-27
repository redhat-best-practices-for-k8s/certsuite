FROM registry.access.redhat.com/ubi9/python-312@sha256:0abad955c2c2e4575dc8987d76b8c49a7dd8e0679063e3118dab03c54528232f

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
