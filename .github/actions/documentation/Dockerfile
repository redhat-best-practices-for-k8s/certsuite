FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:c9e27387a522762a28ece41494070915066b335a47ecf474ebecb08e53e59dbe

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
