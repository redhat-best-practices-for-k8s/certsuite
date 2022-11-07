package securitycontextcontainer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckPod(t *testing.T) {
	runAs := int64(20000)
	allopiv := true
	testCases := []struct {
		testSlice     *provider.Pod
		expectedSlice []PodListcategory
	}{
		{
			testSlice: createPod(runAs, allopiv, "KILL", "MKNOD", "SETUID", "SETGID"),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			testSlice: createPod(runAs, allopiv, "AAA", "MKNOD", "SETUID", "SETGID"),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID4,
			}},
			// its a failed one
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedSlice, CheckPod(tc.testSlice))
	}
}

func createPod(runAs int64, allopiv bool, st1, st2, st3, st4 string) *provider.Pod {
	return &provider.Pod{
		Pod: &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test",
				Namespace: "tnf",
			},
			Spec: corev1.PodSpec{
				HostIPC:     false,
				HostNetwork: false,
				HostPID:     false,
				SecurityContext: &corev1.PodSecurityContext{
					RunAsUser:  &runAs,
					RunAsGroup: &runAs,
					FSGroup:    &runAs,
				},
				Containers: []corev1.Container{
					{
						Name: "test",
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &allopiv,
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability{
									corev1.Capability(st1), corev1.Capability(st2), corev1.Capability(st3),
									corev1.Capability(st4),
								},
							},
							SELinuxOptions: &corev1.SELinuxOptions{
								Level: "s0:c123,c456",
							},
							RunAsNonRoot: &allopiv,
						},
					},
				},
			},
		},
	}
}
