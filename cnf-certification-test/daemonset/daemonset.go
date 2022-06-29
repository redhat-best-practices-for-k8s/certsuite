package daemonset

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/test-network-function/cnf-certification-test/internal/clientsholder"
	appsv1 "k8s.io/api/apps/v1"
	v1 "k8s.io/api/apps/v1"
	v1core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

const (
	timeout = 5 * time.Minute
)

func isDaemonSetReady(status *appsv1.DaemonSetStatus) bool {
	//nolint:gocritic
	return status.DesiredNumberScheduled == status.CurrentNumberScheduled &&
		status.DesiredNumberScheduled == status.NumberAvailable &&
		status.DesiredNumberScheduled == status.NumberReady &&
		status.NumberMisscheduled == 0
}
func WaitDaemonsetReady(namespace, name string, timeout time.Duration) error {
	client := clientsholder.GetClientsHolder().K8sClient
	nodes, err := client.CoreV1().Nodes().List(context.TODO(), metav1.ListOptions{})
	if err != nil {
		return fmt.Errorf("failed to get node list, err:%s", err)
	}

	nodesCount := int32(len(nodes.Items))
	isReady := false
	for start := time.Now(); !isReady && time.Since(start) < timeout; {
		daemonSet, err := client.AppsV1().DaemonSets(namespace).Get(context.Background(), name, metav1.GetOptions{})

		if err != nil {
			return fmt.Errorf("failed to get daemonset, err: %s", err)
		}

		if daemonSet.Status.DesiredNumberScheduled != nodesCount {
			return fmt.Errorf("daemonset DesiredNumberScheduled not equal to number of nodes:%d, please instantiate debug pods on all nodes", nodesCount)
		}

		logrus.Infof("Waiting for (%d) debug pods to be ready: %+v", nodesCount, daemonSet.Status)
		if isDaemonSetReady(&daemonSet.Status) {
			isReady = true
			break
		}

		time.Sleep(time.Duration(timeout.Minutes()))
	}

	if !isReady {
		return errors.New("daemonset debug pods not ready")
	}

	logrus.Infof("All the debug pods are ready.")
	return nil
}

func CreateDaemonSetsTemplate(dsName, namespace, containerName, imageWithVersion string) *v1.DaemonSet {

	dsAnnotations := make(map[string]string)
	dsAnnotations["debug.openshift.io/source-container"] = containerName
	dsAnnotations["openshift.io/scc"] = "node-exporter"
	matchLabels := make(map[string]string)
	matchLabels["name"] = dsName

	var trueBool bool = true
	var zeroInt int64 = 0
	var zeroInt32 int32 = 0
	var preempt = v1core.PreemptLowerPriority
	var tolerationsSeconds int64 = 300
	var hostPathType = v1core.HostPathDirectory

	container := v1core.Container{
		Name:            containerName,
		Image:           imageWithVersion,
		ImagePullPolicy: "Always",
		//Command:         []string{"/bin/sh"},
		SecurityContext: &v1core.SecurityContext{
			Privileged: &trueBool,
			RunAsUser:  &zeroInt,
		},
		Stdin:                  true,
		StdinOnce:              true,
		TerminationMessagePath: "/dev/termination-log",
		TTY:                    true,
		VolumeMounts: []v1core.VolumeMount{
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
			Template: v1core.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: matchLabels,
				},
				Spec: v1core.PodSpec{
					Containers:       []v1core.Container{container},
					PreemptionPolicy: &preempt,
					Priority:         &zeroInt32,
					HostNetwork:      true,
					Tolerations: []v1core.Toleration{
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
					Volumes: []v1core.Volume{
						{
							Name: "host",
							VolumeSource: v1core.VolumeSource{
								HostPath: &v1core.HostPathVolumeSource{
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
func DeleteDaemonSet(daemonSetName, namespace string) error {
	fmt.Printf("Deleting daemon set %s\n", daemonSetName)
	deletePolicy := metav1.DeletePropagationForeground
	client := clientsholder.GetClientsHolder().K8sClient.AppsV1()
	if err := client.DaemonSets(namespace).Delete(context.TODO(), daemonSetName, metav1.DeleteOptions{
		PropagationPolicy: &deletePolicy,
	}); err != nil {
		logrus.Infof("The daemonset (%d) deletion is unsuccessful due to %+v", daemonSetName, err.Error())
	}
	doneCleanUp := false
	for start := time.Now(); !doneCleanUp && time.Since(start) < timeout; {
		client := clientsholder.GetClientsHolder().K8sClient
		pods, err := client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "name=" + daemonSetName})
		if err != nil {
			return fmt.Errorf("failed to get pods, err: %s", err)
		}

		if len(pods.Items) == 0 {
			doneCleanUp = true
			break
		}
		time.Sleep(time.Duration(timeout.Minutes()))
	}

	fmt.Printf("Successfully cleaned up daemon set %s\n", daemonSetName)
	return nil
}

// Check if the daemon set exists
func doesDaemonSetExist(daemonSetName, namespace string) bool {
	client := clientsholder.GetClientsHolder().K8sClient.AppsV1()
	_, err := client.DaemonSets(namespace).Get(context.TODO(), daemonSetName, metav1.GetOptions{})
	if err != nil {
		fmt.Println("Error ocurred for" + err.Error())
	}
	// If the error is not found, that means the daemon set exists
	return err == nil
}
func CreateDaemonSet(daemonSetName, namespace, containerName, imageWithVersion string, timeout time.Duration) (*v1core.PodList, error) {
	rebootDaemonSet := CreateDaemonSetsTemplate(daemonSetName, namespace, containerName, imageWithVersion)
	if doesDaemonSetExist(daemonSetName, namespace) {
		err := DeleteDaemonSet(daemonSetName, namespace)
		if err != nil {
			logrus.Debug("Failed to delete L2discovery daemonset because: %s", err)
		}
	}

	fmt.Printf("Creating daemon set %s\n", daemonSetName)
	client := clientsholder.GetClientsHolder().K8sClient
	_, err := client.AppsV1().DaemonSets(namespace).Create(context.TODO(), rebootDaemonSet, metav1.CreateOptions{})
	if err != nil {
		return nil, err
	}
	WaitDaemonsetReady(namespace, daemonSetName, timeout)

	fmt.Println("Deamon set is ready")

	var ptpPods *v1core.PodList
	ptpPods, err = client.CoreV1().Pods(namespace).List(context.TODO(), metav1.ListOptions{LabelSelector: "name=" + daemonSetName})
	if err != nil {
		return ptpPods, err
	}
	fmt.Printf("Successfully created daemon set %s\n", daemonSetName)
	return ptpPods, nil
}

func partnerRepoDaemonset(daemonSetName, namespace, containerName, imageWithVersion string) {
	dsRunningPods, err := CreateDaemonSet(daemonSetName, namespace, containerName, imageWithVersion, timeout)
	if err != nil {
		logrus.Errorf("Error : +%v\n", err.Error())
	}

	nodeToPodMapping := make(map[string]v1core.Pod)
	for _, dsPod := range dsRunningPods.Items {
		nodeToPodMapping[dsPod.Spec.NodeName] = dsPod
	}
}
