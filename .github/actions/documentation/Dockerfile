FROM registry.access.redhat.com/ubi9/python-39:1-186.1720018722@sha256:145f08b49fabbfec82c9042f6de0d3160c6005ecf700b7fcee3c482e24fd0bf6

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
