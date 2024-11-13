FROM registry.access.redhat.com/ubi9/python-39:9.5@sha256:9a69bc125da09d5cf39497519a955ec958e6618ea6874e46a8c571391dea0393

# Pin versions in pip.
# hadolint ignore=DL3013
COPY frozen-requirements.txt frozen-requirements.txt
RUN pip install --no-cache-dir --require-hashes -r frozen-requirements.txt

COPY entrypoint.sh /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
