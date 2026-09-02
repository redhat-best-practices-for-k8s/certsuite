package checksadapter

import (
	"context"
	"testing"

	"github.com/redhat-best-practices-for-k8s/certsuite/internal/clientsholder"
	"github.com/redhat-best-practices-for-k8s/checks"
	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestProbeExecutorAdapter_Interface(t *testing.T) {
	var _ checks.ProbeExecutor = (*ProbeExecutorAdapter)(nil)
}

func TestProbeExecutorAdapter_ExecCommand_NoContainers(t *testing.T) {
	mock := &clientsholder.MockCommand{
		ExecFunc: func(ctx clientsholder.Context, command string) (string, string, error) {
			assert.Equal(t, "container-00", ctx.GetContainerName())
			assert.Equal(t, "ls", command)
			return "out", "err", nil
		},
	}
	adapter := &ProbeExecutorAdapter{executor: mock}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-ns",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{},
		},
	}

	stdout, stderr, err := adapter.ExecCommand(context.TODO(), pod, "ls")
	assert.NoError(t, err)
	assert.Equal(t, "out", stdout)
	assert.Equal(t, "err", stderr)
	assert.Equal(t, 1, mock.CallCount())
}

func TestProbeExecutorAdapter_ExecCommand_WithContainers(t *testing.T) {
	mock := &clientsholder.MockCommand{
		ExecFunc: func(ctx clientsholder.Context, command string) (string, string, error) {
			assert.Equal(t, "c1", ctx.GetContainerName())
			return "out", "", nil
		},
	}
	adapter := &ProbeExecutorAdapter{executor: mock}
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "test-ns",
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{Name: "c1"},
				{Name: "c2"},
			},
		},
	}

	_, _, err := adapter.ExecCommand(context.TODO(), pod, "ls")
	assert.NoError(t, err)
	assert.Equal(t, 1, mock.CallCount())
}
