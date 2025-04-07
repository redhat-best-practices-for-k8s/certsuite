FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:dd4267b8fae4fbaeefbc5396057cf4acfcc6e4147a5e1c1eb54240e2d6c1b1e6

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
