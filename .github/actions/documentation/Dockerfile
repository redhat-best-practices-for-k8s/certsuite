FROM registry.access.redhat.com/ubi9/python-312@sha256:ff373f4b42b662e99954adea770ca87b4ea963186cc752174ccb94aa08fa702d

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
