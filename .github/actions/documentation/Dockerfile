FROM registry.access.redhat.com/ubi9/python-39:1-186@sha256:5e280686bdb36e2317118f93c3de9e59ce2171dce69a3ddaf241399f576b367b

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
