FROM registry.access.redhat.com/ubi9/python-39:1-182.1715774360@sha256:24a03d125e857440f74b445d8deb2c18775dd7c3c4373ce1662a00cf97ef8e25

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
