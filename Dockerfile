FROM registry.access.redhat.com/ubi8/ubi:latest AS build

ARG OPENSHIFT_VERSION
ENV OPENSHIFT_VERSION=${OPENSHIFT_VERSION}
ENV TNF_DIR=/usr/tnf
ENV TNF_SRC_DIR=${TNF_DIR}/tnf-src
ENV TNF_BIN_DIR=${TNF_DIR}/cnf-certification-test

RUN mkdir ${TNF_DIR}

ENV TEMP_DIR=/tmp

# Install dependencies
RUN yum install -y gcc make wget

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
