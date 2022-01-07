FROM registry.access.redhat.com/ubi8/ubi:latest AS build
ARG TNF_PARTNER_DIR=/usr/tnf-partner

ENV TNF_PARTNER_SRC_DIR=$TNF_PARTNER_DIR/src

ENV OPENSHIFT_VERSION=4.7.7

ENV TNF_DIR=/usr/tnf
ENV TNF_SRC_DIR=${TNF_DIR}/tnf-src
ENV TNF_BIN_DIR=${TNF_DIR}/test-network-function

ENV TEMP_DIR=/tmp

# Install dependencies
RUN yum install -y gcc git jq make wget

# Install Go binary
ENV GO_DL_URL="https://golang.org/dl"
ENV GO_BIN_TAR="go1.17.4.linux-amd64.tar.gz"
ENV GO_BIN_URL_x86_64=${GO_DL_URL}/${GO_BIN_TAR}
ENV GOPATH="/root/go"
RUN if [[ "$(uname -m)" -eq "x86_64" ]] ; then \
        wget --directory-prefix=${TEMP_DIR} ${GO_BIN_URL_x86_64} && \
            rm -rf /usr/local/go && \
            tar -C /usr/local -xzf ${TEMP_DIR}/${GO_BIN_TAR}; \
     else \
         echo "CPU architecture not supported" && exit 1; \
     fi

# Install oc binary
ENV OC_BIN_TAR="openshift-client-linux.tar.gz"
ENV OC_DL_URL="https://mirror.openshift.com/pub/openshift-v4/clients/ocp"/${OPENSHIFT_VERSION}/${OC_BIN_TAR}
ENV OC_BIN="/usr/local/oc/bin"
RUN wget --directory-prefix=${TEMP_DIR} ${OC_DL_URL} && \
    mkdir -p ${OC_BIN} && \
    tar -C ${OC_BIN} -xzf ${TEMP_DIR}/${OC_BIN_TAR} && \
    chmod a+x ${OC_BIN}/oc

# Add go and oc binary directory to $PATH
ENV PATH=${PATH}:"/usr/local/go/bin":${GOPATH}/"bin"

# Git identifier to checkout
ARG TNF_VERSION
ARG TNF_SRC_URL=https://github.com/test-network-function/test-network-function
ARG GIT_CHECKOUT_TARGET=$TNF_VERSION

# Git identifier to checkout for partner
ARG TNF_PARTNER_VERSION
ARG TNF_PARTNER_SRC_URL=https://github.com/test-network-function/cnf-certification-test-partner
ARG GIT_PARTNER_CHECKOUT_TARGET=$TNF_PARTNER_VERSION

# Clone the TNF source repository and checkout the target branch/tag/commit
RUN git clone --no-single-branch --depth=1 ${TNF_SRC_URL} ${TNF_SRC_DIR}
RUN git -C ${TNF_SRC_DIR} fetch origin ${GIT_CHECKOUT_TARGET}
RUN git -C ${TNF_SRC_DIR} checkout ${GIT_CHECKOUT_TARGET}

# Clone the partner source repository and checkout the target branch/tag/commit
RUN git clone --no-single-branch --depth=1 ${TNF_PARTNER_SRC_URL} ${TNF_PARTNER_SRC_DIR}
RUN git -C ${TNF_PARTNER_SRC_DIR} fetch origin ${GIT_PARTNER_CHECKOUT_TARGET}
RUN git -C ${TNF_PARTNER_SRC_DIR} checkout ${GIT_PARTNER_CHECKOUT_TARGET}

# Build TNF binary
WORKDIR ${TNF_SRC_DIR}

# golangci-lint
RUN make install-lint 

# TODO: RUN make install-tools
RUN make install-tools && \
	make mocks && \
	make update-deps && \
	make build-cnf-tests-debug

#  Extract what's needed to run at a seperate location
RUN mkdir ${TNF_BIN_DIR} && \
	cp run-cnf-suites.sh ${TNF_DIR} && \
    mkdir ${TNF_DIR}/script && \
    cp script/results.html ${TNF_DIR}/script && \
	# copy all JSON files to allow tests to run
	cp --parents `find -name \*.json*` ${TNF_DIR} && \
  # copy all go template files to allow tests to run
	cp --parents `find -name \*.gotemplate*` ${TNF_DIR} && \
	cp test-network-function/test-network-function.test ${TNF_BIN_DIR}

WORKDIR ${TNF_DIR}

RUN ln -s ${TNF_DIR}/config/testconfigure.yml ${TNF_DIR}/test-network-function/testconfigure.yml

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

# Copy the state into a new flattened image to reduce size.
# TODO run as non-root
FROM scratch
ARG TNF_PARTNER_DIR=/usr/tnf-partner
COPY --from=build / /
ENV TNF_CONFIGURATION_PATH=/usr/tnf/config/tnf_config.yml
ENV KUBECONFIG=/usr/tnf/kubeconfig/config
ENV TNF_PARTNER_SRC_DIR=$TNF_PARTNER_DIR/src
ENV PATH="/usr/local/oc/bin:${PATH}"
WORKDIR /usr/tnf
ENV SHELL=/bin/bash
CMD ["./run-cnf-suites.sh", "-o", "claim", "-f", "diagnostic"]
