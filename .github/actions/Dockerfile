FROM registry.access.redhat.com/ubi8/ubi

RUN dnf install -y python39
RUN pip3 install mkdocs
RUN pip3 install mkdocs-material
RUN pip3 install markdown-include

COPY entrypoint.sh /entrypoint.sh
RUN chmod +x /entrypoint.sh

ENTRYPOINT ["/entrypoint.sh"]
