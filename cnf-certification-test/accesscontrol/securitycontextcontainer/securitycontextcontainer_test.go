package securitycontextcontainer

/*import (
	"testing"

	corev1 "k8s.io/api/core/v1"

	"github.com/stretchr/testify/assert"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

func TestRemoveEmptyStrings(t *testing.T) {
	var runAs *int64
	*runAs = 2000
	var allopiv *bool
	*allopiv = true
	testCases := []struct {
		testSlice     *provider.Pod
		expectedSlice []string
	}{
		{
			testSlice: &provider.Pod{
				Pod: &corev1.Pod{
					ObjectMeta: metav1.ObjectMeta{},
					Spec: corev1.PodSpec{
						HostIPC:     false,
						HostNetwork: false,
						HostPID:     false,
						SecurityContext: &corev1.PodSecurityContext{
							RunAsUser:  runAs,
							RunAsGroup: runAs,
							FSGroup:    runAs,
						},
						Containers: []corev1.Container{
							corev1.Container{
								SecurityContext: &corev1.SecurityContext{
									AllowPrivilegeEscalation: allopiv,
									Capabilities: &corev1.Capabilities{
										Drop: []corev1.Capability{
											"KILL", "MKNOD", "SETUID", "SETGID",
										},
									},
									SELinuxOptions: &corev1.SELinuxOptions{
										Level: "s0:c123,c456",
									},
								},
							},
						},
					},
				},
			},
			expectedSlice: nil,
		},
	}

	for _, tc := range testCases {
		assert.Equal(t, tc.expectedSlice, CheckPod(tc.testSlice))
	}
}
*/
