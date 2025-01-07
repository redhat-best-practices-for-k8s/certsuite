FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:dde5068ba8fd81eef41a9128ddad2b99e2e02e5869ffe925c605b609a0ad7bbb

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
