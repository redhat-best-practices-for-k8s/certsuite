FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:9c3fbaa5af85700cd2a1c298332d8c7edbf7cca02f02b3977bf67752f141044b

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
