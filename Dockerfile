FROM registry.access.redhat.com/ubi8/ubi:latest AS build

ENV TNF_DIR=/usr/tnf
ENV TNF_SRC_DIR=${TNF_DIR}/tnf-src
ENV TNF_BIN_DIR=${TNF_DIR}/cnf-certification-test

RUN mkdir ${TNF_DIR}

ENV TEMP_DIR=/tmp

# Install dependencies
RUN yum install -y gcc make wget

# Install Go binary and set the PATH 
ENV GO_DL_URL="https://golang.org/dl"
ENV GO_BIN_TAR="go1.19.4.linux-amd64.tar.gz"
ENV GO_BIN_URL_x86_64=${GO_DL_URL}/${GO_BIN_TAR}
ENV GOPATH="/root/go"
RUN if [[ "$(uname -m)" -eq "x86_64" ]] ; then \
        wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_x86_64} && \
            rm -rf /usr/local/go && \
            tar -C /usr/local -xzf ${TEMP_DIR}/${GO_BIN_TAR}; \
     else \
         echo "CPU architecture not supported" && exit 1; \
     fi

ENV PATH=${PATH}:"/usr/local/go/bin":${GOPATH}/"bin"

# Download operator-sdk binary
ENV OPERATOR_SDK_DL_URL=https://github.com/operator-framework/operator-sdk/releases/download/v1.26.0 \
    OSDK_BIN="/usr/local/osdk/bin"
RUN mkdir -p ${OSDK_BIN} && \
    curl --location --remote-name ${OPERATOR_SDK_DL_URL}/operator-sdk_linux_amd64 && \
    mv operator-sdk_linux_amd64 ${OSDK_BIN}/operator-sdk && \
    chmod +x ${OSDK_BIN}/operator-sdk

# Copy all of the files into the source directory and then switch contexts
COPY . ${TNF_SRC_DIR}
WORKDIR ${TNF_SRC_DIR}

RUN make install-tools && \
	make build-cnf-tests

#  Extract what's needed to run at a seperate location
RUN mkdir ${TNF_BIN_DIR} && \
	cp run-cnf-suites.sh ${TNF_DIR} && \
    mkdir ${TNF_DIR}/script && \
    cp script/results.html ${TNF_DIR}/script && \
	# copy all JSON files to allow tests to run
	cp --parents `find -name \*.json*` ${TNF_DIR} && \
	cp cnf-certification-test/cnf-certification-test.test ${TNF_BIN_DIR} && \
    # copy all of the chaos-test-files
    mkdir -p ${TNF_DIR}/cnf-certification-test/chaostesting && \
    cp -a cnf-certification-test/chaostesting/chaos-test-files ${TNF_DIR}/cnf-certification-test/chaostesting && \
    # copy the rhcos_version_map
    mkdir -p ${TNF_DIR}/cnf-certification-test/platform/operatingsystem/files && \
    cp cnf-certification-test/platform/operatingsystem/files/rhcos_version_map ${TNF_DIR}/cnf-certification-test/platform/operatingsystem/files/rhcos_version_map

# Switch contexts back to the root TNF directory
WORKDIR ${TNF_DIR}

# Remove most of the build artefacts
RUN yum remove -y gcc git wget && \
	yum clean all && \
	rm -rf ${TNF_SRC_DIR} && \
	rm -rf ${TEMP_DIR} && \
	rm -rf /root/.cache && \
	rm -rf /root/go/pkg && \
	rm -rf /root/go/src && \
	rm -rf /usr/lib/golang/pkg && \
	rm -rf /usr/lib/golang/src

FROM quay.io/testnetworkfunction/oct:latest AS db

# Copy the state into a new flattened image to reduce size.
# TODO run as non-root
FROM registry.access.redhat.com/ubi8/ubi-minimal:latest

# Copy all of the necessary files over from the TNF_DIR
ENV TNF_DIR=/usr/tnf
COPY --from=build ${TNF_DIR} ${TNF_DIR}

# Add operatorsdk binary to image
COPY --from=build ${OSDK_BIN} ${OSDK_BIN}

# Update the CNF containers, helm charts and operators DB
ENV TNF_OFFLINE_DB=/usr/offline-db \
    OCT_DB_PATH=/usr/oct/cmd/tnf/fetch
COPY --from=db ${OCT_DB_PATH} ${TNF_OFFLINE_DB}

ENV TNF_CONFIGURATION_PATH=/usr/tnf/config/tnf_config.yml \
    KUBECONFIG=/usr/tnf/kubeconfig/config \
    PATH="/usr/local/osdk/bin:${PATH}"
WORKDIR ${TNF_DIR}
ENV SHELL=/bin/bash
CMD ["./run-cnf-suites.sh", "-o", "claim", "-f", "diagnostic"]
