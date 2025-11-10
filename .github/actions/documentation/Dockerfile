FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:8093b0752e9cce0019c0708f55a032c158d70fefe31a2dcb13d179311e356e5e

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
