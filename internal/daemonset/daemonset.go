package daemonset

import (
	"context"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	"github.com/test-network-function/cnf-certification-test/pkg/provider"
	v1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	imageWithVersion                   = "quay.io/testnetworkfunction/debug-partner:latest"
	timeout                            = 5 * time.Minute
	daemonsetDeletionCheckRetryInteval = 5 * time.Second
	nodeExporter                       = "node-exporter"
	containerName                      = "container-00"
	debug                              = "debug"
)

// Delete daemon set
func DeleteDaemonSet(daemonSetName, namespace string) error {
	logrus.Infof("Deleting daemon set %s", daemonSetName)
	deletePolicy := metav1.DeletePropagationForeground
	client := clientsholder.GetClientsHolder().K8sClient
	err := client.AppsV1().DaemonSets(namespace).Delete(context.TODO(), daemonSetName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy})
	if err != nil {
		return fmt.Errorf("daemonset %s deletion failed: %w", daemonSetName, err)
	}
	allPodsRemoved := false
	for start := time.Now(); time.Since(start) < timeout; {
		pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "test-network-function.com/app=debug"})
		if err != nil {
			return fmt.Errorf("failed to get pods, err: %s", err)
		}

		if len(pods.Items) == 0 {
			allPodsRemoved = true
			break
		}
		allPodsRemoved = false
		time.Sleep(daemonsetDeletionCheckRetryInteval)
	}

	if !allPodsRemoved {
		return fmt.Errorf("timeout waiting for daemonset's pods to be deleted")
	}
	logrus.Infof("Successfully cleaned up daemon set %s", daemonSetName)
	return nil
}

// Check if the daemon set exists
func doesDaemonSetExist(daemonSetName, namespace string) bool {
	client := clientsholder.GetClientsHolder().K8sClient.AppsV1()
	_, err := client.DaemonSets(namespace).Get(context.TODO(), daemonSetName, metav1.GetOptions{})
	if err != nil {
		logrus.Infof("Error occurred checking for Daemonset to exist: %s", err)
		return false
	}
	return true
}

//nolint:funlen
func CreateDaemonSetsTemplate(dsName, namespace, containerName, imageWithVersion string) *v1.DaemonSet {
	dsAnnotations := make(map[string]string)
	dsAnnotations["debug.openshift.io/source-container"] = containerName
	dsAnnotations["openshift.io/scc"] = nodeExporter
	matchLabels := make(map[string]string)
	matchLabels["name"] = dsName
	matchLabels["test-network-function.com/app"] = debug

	var runAsPrivileged = true
	var zeroInt int64
	var zeroInt32 int32
	var preempt = corev1.PreemptLowerPriority
	var tolerationsSeconds int64 = 300
	var hostPathType = corev1.HostPathDirectory

	container := corev1.Container{
		Name:            containerName,
		Image:           imageWithVersion,
		ImagePullPolicy: corev1.PullIfNotPresent,
		SecurityContext: &corev1.SecurityContext{
			Privileged: &runAsPrivileged,
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
					HostPID:          true,
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

// Create daemon set
func CreateDaemonSet(daemonSetName, namespace, containerName, imageWithVersion string, timeout time.Duration) (*corev1.PodList, error) {
	aDaemonSet := CreateDaemonSetsTemplate(daemonSetName, namespace, containerName, imageWithVersion)
	if doesDaemonSetExist(daemonSetName, namespace) {
		err := DeleteDaemonSet(daemonSetName, namespace)
		if err != nil {
			return nil, err
		}
	}

	logrus.Infof("Creating daemon set %s", daemonSetName)
	client := clientsholder.GetClientsHolder().K8sClient
	_, err := client.AppsV1().DaemonSets(namespace).Create(context.TODO(), aDaemonSet, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	err = provider.WaitDebugPodsReady()
	if err != nil {
		return nil, err
	}

	logrus.Infof("Daemonset is ready")

	var daemonsetPods *corev1.PodList
	daemonsetPods, err = client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "test-network-function.com/app=debug"})
	if err != nil {
		return nil, err
	}
	logrus.Infof("Successfully created daemon set %s", daemonSetName)
	return daemonsetPods, nil
}

// Deploy daemon set on repo partner
func DeployPartnerTestDaemonset() error {
	_, err := CreateDaemonSet(provider.DaemonSetName, provider.DaemonSetNamespace, containerName, imageWithVersion, timeout)
	if err != nil {
		logrus.Errorf("Error deploying partner daemonset %s", err)
		return err
	}
	return nil
}
