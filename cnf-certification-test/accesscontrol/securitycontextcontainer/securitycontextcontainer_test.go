package securitycontextcontainer

import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

//nolint:funlen
func TestCheckPod(t *testing.T) {
	runAs := int64(20000)
	runAs2 := int64(1000)
	testCases := []struct {
		testSlice     *provider.Pod
		expectedSlice []PodListcategory
	}{
		{
			// Category 1 - pass
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			// Category 1 no UIOD0 - pass
			testSlice: createPod(runAs2, true, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1NoUID0,
			}},
		},
		{
			// Category 2 - pass
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{"NET_ADMIN", "NET_RAW"}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID2,
			}},
		},
		{
			// category 3 - pass
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{"IPC_LOCK", "NET_ADMIN", "NET_RAW"}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID3,
			}},
		},
		{
			// Fail due to required drop capabilities missing
			testSlice: createPod(runAs2, false, true, []corev1.Capability{"SYS_TIME", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID4,
			}},
		},
		{
			// Category 1 - passing with extra drop capabilities
			testSlice: createPod(runAs2, false, true, []corev1.Capability{"SYS_TIME", "KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			// Category 1 - passing with ALL drop capabilities
			testSlice: createPod(runAs2, false, true, []corev1.Capability{"ALL"}, []corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			// Category 1 - pass with no privilege escalation
			testSlice: createPod(runAs, false, false, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListcategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedSlice, CheckPod(tc.testSlice))
	}
}

func createPod(runAs int64, runAsNonRootParam, allowPrivilegeEscalationParam bool, drop, add []corev1.Capability) *provider.Pod {
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
							AllowPrivilegeEscalation: &allowPrivilegeEscalationParam,
							Capabilities: &corev1.Capabilities{
								Drop: drop,
								Add:  add,
							},
							SELinuxOptions: &corev1.SELinuxOptions{
								Level: "s0:c123,c456",
							},
							RunAsNonRoot: &runAsNonRootParam,
						},
					},
				},
			},
		},
	}
}
