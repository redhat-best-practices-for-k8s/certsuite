FROM registry.access.redhat.com/ubi9/python-39:9.7@sha256:b9cab177ccea1acd7b1df9a23dcfa280177c7c66615dd7567036adb7820f49b3

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
