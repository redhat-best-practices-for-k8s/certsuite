FROM registry.access.redhat.com/ubi9/python-314@sha256:fd506034c95cb5917c80fe480937e1fc17564ab7812f3fded84899ed0cd6e9e4

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
