FROM registry.access.redhat.com/ubi9/python-312@sha256:12662c50b0ce52c5a5f43724c9c588c6cbe9be98521f4e3f4112ab01a7becbe5

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
