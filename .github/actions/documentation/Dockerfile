FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:f0706d250b9712a7544230a6ac8507ac46cfddd14e8016ab5f30bbaea9bf3a80

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
