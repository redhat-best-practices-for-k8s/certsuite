package daemonset

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	"github.com/test-network-function/cnf-certification-test/pkg/tnf"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	daemonSetName    = "partner-repo"
	namespace        = "default"
	containerName    = "container-00"
	imageWithVersion = "quay.io/testnetworkfunction/debug-partner:latest"
	timeout          = 5 * time.Minute
	nodeExporter     = "node-exporter"
)

//nolint:funlen
func CreateDaemonSetsTemplate(dsName, namespace, containerName, imageWithVersion string) *v1.DaemonSet {
	dsAnnotations := make(map[string]string)
	dsAnnotations["debug.openshift.io/source-container"] = containerName
	dsAnnotations["openshift.io/scc"] = nodeExporter
	matchLabels := make(map[string]string)
	matchLabels["name"] = dsName

	var trueBool = true
	var zeroInt int64
	var zeroInt32 int32
	var preempt = corev1.PreemptLowerPriority
	var tolerationsSeconds int64 = 300
	var hostPathType = corev1.HostPathDirectory

	container := corev1.Container{
		Name:            containerName,
		Image:           imageWithVersion,
		ImagePullPolicy: "IfNotPresent",
		SecurityContext: &corev1.SecurityContext{
			Privileged: &trueBool,
			RunAsUser:  &zeroInt,
		},
		Stdin:                  true,
		StdinOnce:              true,
		TerminationMessagePath: "/dev/termination-log",
		TTY:                    true,
		VolumeMounts: []corev1.VolumeMount{
			{
				MountPath: "/host",
				Name:      "host",
			},
		},
	}
	return &v1.DaemonSet{

		ObjectMeta: metav1.ObjectMeta{
			Name:        dsName,
			Namespace:   namespace,
			Annotations: dsAnnotations,
		},
		Spec: v1.DaemonSetSpec{
			Selector: &metav1.LabelSelector{
				MatchLabels: matchLabels,
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: corev1.PodSpec{
					Containers:       []corev1.Container{container},
					PreemptionPolicy: &preempt,
					Priority:         &zeroInt32,
					HostNetwork:      true,
					Tolerations: []corev1.Toleration{
						{
							Effect:            "NoExecute",
							Key:               "node.kubernetes.io/not-ready",
							Operator:          "Exists",
							TolerationSeconds: &tolerationsSeconds,
						},
						{
							Effect:            "NoExecute",
							Key:               "node.kubernetes.io/unreachable",
							Operator:          "Exists",
							TolerationSeconds: &tolerationsSeconds,
						},
						{
							Effect: "NoSchedule",
							Key:    "node-role.kubernetes.io/master",
						},
					},
					Volumes: []corev1.Volume{
						{
							Name: "host",
							VolumeSource: corev1.VolumeSource{
								HostPath: &corev1.HostPathVolumeSource{
									Path: "/",
									Type: &hostPathType,
								},
							},
						},
					},
				},
			},
		},
	}
}

// Delete dameno set
func DeleteDaemonSet(daemonSetName, namespace string) error {
	tnf.ClaimFilePrintf("Deleting daemon set %s\n", daemonSetName)
	deletePolicy := metav1.DeletePropagationForeground
	client := clientsholder.GetClientsHolder().K8sClient.AppsV1()
	if err := client.DaemonSets(namespace).Delete(context.TODO(), daemonSetName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		tnf.ClaimFilePrintf("The daemonset (%s) deletion is unsuccessful due to %s", daemonSetName, err)
	}
	doneCleanUp := false
	for start := time.Now(); !doneCleanUp && time.Since(start) < timeout; {
		client := clientsholder.GetClientsHolder().K8sClient
		pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "name=" + daemonSetName})
		if err != nil {
			return fmt.Errorf("failed to get pods, err: %s", err)
		}
		//nolint:ineffassign
		if len(pods.Items) == 0 {
			doneCleanUp = true
			break
		}
		time.Sleep(time.Duration(timeout.Minutes()))
	}

	tnf.ClaimFilePrintf("Successfully cleaned up daemon set %s\n", daemonSetName)
	return nil
}

// Check if the daemon set exists
func doesDaemonSetExist(daemonSetName, namespace string) bool {
	client := clientsholder.GetClientsHolder().K8sClient.AppsV1()
	_, err := client.DaemonSets(namespace).Get(context.TODO(), daemonSetName, metav1.GetOptions{})
	if err != nil {
		tnf.ClaimFilePrintf("Error occurred checking for Daemonset to exist: " + err.Error())
	}
	// If the error is not found, that means the daemon set exists
	return err == nil
}

// Create daemon set
func CreateDaemonSet(daemonSetName, namespace, containerName, imageWithVersion string, timeout time.Duration) (*corev1.PodList, error) {
	rebootDaemonSet := CreateDaemonSetsTemplate(daemonSetName, namespace, containerName, imageWithVersion)
	if doesDaemonSetExist(daemonSetName, namespace) {
		err := DeleteDaemonSet(daemonSetName, namespace)
		if err != nil {
			tnf.ClaimFilePrintf("Failed to delete debug daemonset because: %s", err)
		}
	}

	tnf.ClaimFilePrintf("Creating daemon set %s\n", daemonSetName)
	client := clientsholder.GetClientsHolder().K8sClient
	_, err := client.AppsV1().DaemonSets(namespace).Create(context.TODO(), rebootDaemonSet, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	err = provider.WaitDebugPodsReady()
	if err != nil {
		return nil, err
	}

	tnf.ClaimFilePrintf("Daemonset is ready")

	var ptpPods *corev1.PodList
	ptpPods, err = client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "name=" + daemonSetName})
	if err != nil {
		return ptpPods, err
	}
	tnf.ClaimFilePrintf("Successfully created daemon set %s\n", daemonSetName)
	return ptpPods, nil
}

// Deploy daemon set on repo partner
func PartnerRepoDaemonset() map[string]corev1.Pod {
	dsRunningPods, err := CreateDaemonSet(daemonSetName, namespace, containerName, imageWithVersion, timeout)
	if err != nil {
		logrus.Errorf("Error : +%v\n", err.Error())
	}

	nodeToPodMapping := make(map[string]corev1.Pod)
	//nolint:gocritic
	for _, dsPod := range dsRunningPods.Items {
		nodeToPodMapping[dsPod.Spec.NodeName] = dsPod
	}
	return nodeToPodMapping
}
