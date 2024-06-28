package checksdb

import (
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestString(t *testing.T) {
	cr := CheckResult("passed")
	assert.Equal(t, "passed", cr.String())
}

func TestNewCheck(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	assert.NotNil(t, check)

	assert.Equal(t, "myID", check.ID)
	assert.Equal(t, []string{"label1", "label2"}, check.Labels)
}

func TestSetAbortChan(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})
	abortChan := make(chan string)

	check.SetAbortChan(abortChan)

	assert.Equal(t, abortChan, check.abortChan)
}

func TestGetLogs(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	logs := check.GetLogs()

	assert.NotNil(t, logs)
}

func TestGetLogger(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	logger := check.GetLogger()

	assert.NotNil(t, logger)
}

func TestWithCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.CheckFn)

	check.Error = errors.New("this is an error")
	check.WithCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.CheckFn)
	assert.Equal(t, "this is an error", check.Error.Error())
}

func TestWithBeforeCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithBeforeCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.BeforeCheckFn)

	check.Error = errors.New("this is an error")
	check.WithBeforeCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.BeforeCheckFn)
	assert.Equal(t, "this is an error", check.Error.Error())
}

func TestWithAfterCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithAfterCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.AfterCheckFn)

	check.Error = errors.New("this is an error")
	check.WithAfterCheckFn(func(check *Check) error {
		return nil
	})

	assert.NotNil(t, check.AfterCheckFn)
	assert.Equal(t, "this is an error", check.Error.Error())
}

func TestWithSkipCheckFn(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithSkipCheckFn(func() (skip bool, reason string) {
		return false, ""
	})

	assert.Len(t, check.SkipCheckFns, 1)

	check.WithSkipCheckFn(func() (skip bool, reason string) {
		return false, ""
	})

	assert.Len(t, check.SkipCheckFns, 2)
}

func TestWithSkipModeAny(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	// Test the default value, which is SkipModeAny
	assert.Equal(t, SkipModeAny, check.SkipMode)

	check.WithSkipModeAny()

	assert.Equal(t, SkipModeAny, check.SkipMode)
}

func TestWithSkipModeAll(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	// Test the default value, which is SkipModeAny
	assert.Equal(t, SkipModeAny, check.SkipMode)

	check.WithSkipModeAll()

	assert.Equal(t, SkipModeAll, check.SkipMode)
}

func TestWithTimeout(t *testing.T) {
	check := NewCheck("myID", []string{"label1", "label2"})

	check.WithTimeout(10)

	assert.Equal(t, time.Duration(10), check.Timeout)
}
