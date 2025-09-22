FROM registry.access.redhat.com/ubi9/python-39:9.6@sha256:842eeb962621099270f52246499b16c7da5204b76d4a33c94d59339061404827

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
