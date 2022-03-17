package tnf

import (
	"fmt"

	"github.com/onsi/ginkgo/v2"
	"github.com/onsi/gomega"
	"github.com/sirupsen/logrus"
)

// ClaimFilePrintf prints to claim and junit report files.
func ClaimFilePrintf(format string, args ...interface{}) {
	message := fmt.Sprintf(format+"\n", args...)
	_, err := ginkgo.GinkgoWriter.Write([]byte(message))
	if err != nil {
		logrus.Errorf("Ginkgo writer could not write msg '%s' because: %s", message, err)
	} else {
		logrus.Trace(message)
	}
}

//go:generate moq -out ginkgofuncs_moq.go . GinkgoFuncs
type GinkgoFuncs interface {
	GinkgoBy(text string, callback ...func())
	GinkgoFail(message string, callerSkip ...int)
	GinkgoSkip(message string, callerSkip ...int)
	GinkgoAbortSuite(message string, callerSkip ...int)
}

func NewGinkgoWrapper() GinkgoFuncs {
	return &GinkgoWrapper{}
}

type GinkgoWrapper struct{}

func (gw *GinkgoWrapper) GinkgoBy(text string, callback ...func()) {
	ginkgo.By(text, callback...)
}
func (gw *GinkgoWrapper) GinkgoFail(message string, callerSkip ...int) {
	ginkgo.Fail(message, callerSkip...)
}
func (gw *GinkgoWrapper) GinkgoSkip(message string, callerSkip ...int) {
	ginkgo.Skip(message, callerSkip...)
}
func (gw *GinkgoWrapper) GinkgoAbortSuite(message string, callerSkip ...int) {
	ginkgo.AbortSuite(message, callerSkip...)
}

//go:generate moq -out gomegafuncs_moq.go . GomegaFuncs
type GomegaFuncs interface {
	GomegaExpectStringNotEmpty(incomingStr string)
	GomegaExpectSliceBeNil(incomingSlice []string)
}

func NewGomegaWrapper() GomegaFuncs {
	return &GomegaWrapper{}
}

type GomegaWrapper struct{}

func (gw *GomegaWrapper) GomegaExpectStringNotEmpty(incomingStr string) {
	gomega.Expect(incomingStr).ToNot(gomega.BeEmpty())
}

func (gw *GomegaWrapper) GomegaExpectSliceBeNil(incomingSlice []string) {
	gomega.Expect(incomingSlice).To(gomega.BeNil())
}
