FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:9cea5d0cbe82176d5f21781f440ddf077f03beff304f051aa698640731404db7

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
