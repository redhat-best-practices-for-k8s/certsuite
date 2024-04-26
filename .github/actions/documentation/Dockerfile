FROM registry.access.redhat.com/ubi9/python-39:1-172.1712567222@sha256:c96f839e927c52990143df4efb2872946fcd5de9e1ed2014947bb2cf3084c27a

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
