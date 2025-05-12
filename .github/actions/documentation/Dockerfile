FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:2214c67e2c766357da98cc9ba6c88525495247d7745ba2b8ab5d365f98ccd59d

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
