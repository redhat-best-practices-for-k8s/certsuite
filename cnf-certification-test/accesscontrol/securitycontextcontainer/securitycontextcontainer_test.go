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

	testCases := []struct {
		testSlice     *provider.Pod
		expectedSlice []PodListcategory
	}{
		{
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"},[]corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			testSlice: createPod(runAs, true, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"},[]corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1NoUID0,
			}},
		},
				{
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{"NET_ADMIN", "NET_RAW"}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID2,
			}},
		},
		{
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{"IPC_LOCK", "NET_ADMIN", "NET_RAW"}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID3,
			}},
		},
		{
			testSlice: createPod(runAs, true, true, []corev1.Capability{"AAA", "MKNOD", "SETUID", "SETGID"},[]corev1.Capability{}),
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

func createPod(runAs int64, RunAsNonRootParam,AllowPrivilegeEscalationParam  bool, drop, add []corev1.Capability) *provider.Pod {
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
							AllowPrivilegeEscalation: &AllowPrivilegeEscalationParam,
							Capabilities: &corev1.Capabilities{
								Drop: []corev1.Capability(drop),
								Add: []corev1.Capability(add),
							},
							SELinuxOptions: &corev1.SELinuxOptions{
								Level: "s0:c123,c456",
							},
							RunAsNonRoot: &RunAsNonRootParam,
						},
					},
				},
			},
		},
	}
}
