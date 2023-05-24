package securitycontextcontainer

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestCheckPod(t *testing.T) {
	runAs := int64(20000)
	runAs2 := int64(1000)
	testCases := []struct {
		testSlice     *provider.Pod
		expectedSlice []PodListCategory
	}{
		{
			// Category 1 - pass
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			// Category 1 no UIOD0 - pass
			testSlice: createPod(runAs2, true, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1NoUID0,
			}},
		},
		{
			// Category 2 - pass
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{"NET_ADMIN", "NET_RAW"}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID2,
			}},
		},
		{
			// category 3 - pass
			testSlice: createPod(runAs, false, true, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{"IPC_LOCK", "NET_ADMIN", "NET_RAW"}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID3,
			}},
		},
		{
			// Fail due to required drop capabilities missing
			testSlice: createPod(runAs2, false, true, []corev1.Capability{"SYS_TIME", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID4,
			}},
		},
		{
			// Category 1 - passing with extra drop capabilities
			testSlice: createPod(runAs2, false, true, []corev1.Capability{"SYS_TIME", "KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			// Category 1 - passing with ALL drop capabilities
			testSlice: createPod(runAs2, false, true, []corev1.Capability{"ALL"}, []corev1.Capability{}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			// Category 1 - pass with no privilege escalation
			testSlice: createPod(runAs, false, false, []corev1.Capability{"KILL", "MKNOD", "SETUID", "SETGID"}, []corev1.Capability{}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			}},
		},
		{
			testSlice: createPod2Containers(runAs2, false, true, []corev1.Capability{"ALL"}, []corev1.Capability{}),
			expectedSlice: []PodListCategory{{
				Containername: "test",
				Podname:       "test",
				NameSpace:     "tnf",
				Category:      CategoryID1,
			},
				{
					Containername: "test2",
					Podname:       "test",
					NameSpace:     "tnf",
					Category:      CategoryID4,
				},
			},
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

func createPod2Containers(runAs int64, runAsNonRootParam, allowPrivilegeEscalationParam bool, drop, add []corev1.Capability) *provider.Pod {
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
					{
						Name: "test2",
						SecurityContext: &corev1.SecurityContext{
							AllowPrivilegeEscalation: &allowPrivilegeEscalationParam,
							Capabilities: &corev1.Capabilities{
								Drop: drop,
								Add:  add,
							},
							RunAsNonRoot: &runAsNonRootParam,
						},
					},
				},
			},
		},
	}
}

func TestAllVolumeAllowed(t *testing.T) {
	type args struct {
		volumes []corev1.Volume
	}
	tests := []struct {
		name   string
		args   args
		wantR1 OkNok
		wantR2 OkNok
	}{
		{
			name: "test1",
			args: args{
				volumes: []corev1.Volume{
					{
						Name: "test1",
						VolumeSource: corev1.VolumeSource{
							ConfigMap: &corev1.ConfigMapVolumeSource{
								LocalObjectReference: corev1.LocalObjectReference{
									Name: "test1",
								},
							},
						},
					},
				},
			},
			wantR1: OK,
			wantR2: NOK,
		},
		{
			name: "test2",
			args: args{
				volumes: []corev1.Volume{
					{
						Name: "test2",
						VolumeSource: corev1.VolumeSource{
							GitRepo: &corev1.GitRepoVolumeSource{
								Repository: "test2",
							},
						},
					},
				},
			},
			wantR1: NOK,
			wantR2: NOK,
		},

		{
			name: "test3",
			args: args{
				volumes: []corev1.Volume{
					{
						Name: "test3",
						VolumeSource: corev1.VolumeSource{
							HostPath: &corev1.HostPathVolumeSource{
								Path: "test3",
							},
						},
					},
				},
			},
			wantR1: NOK,
			wantR2: OK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotR1, gotR2 := AllVolumeAllowed(tt.args.volumes)
			if gotR1 != tt.wantR1 {
				t.Errorf("AllVolumeAllowed() gotR1 = %v, want %v", gotR1, tt.wantR1)
			}
			if gotR2 != tt.wantR2 {
				t.Errorf("AllVolumeAllowed() gotR2 = %v, want %v", gotR2, tt.wantR2)
			}
		})
	}
}
