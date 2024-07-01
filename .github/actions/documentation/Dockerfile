FROM registry.access.redhat.com/ubi9/python-39:1-186.1719562233@sha256:c949cb81a6f72ca157501949d07095e87b3c8f1a9ff5a69ff0a858faac21217b

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
