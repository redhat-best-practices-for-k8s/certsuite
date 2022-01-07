# proof of concept bash script to execute any commands in running pods
set -x
NAMESPACE=$1
POD_NAME=$2
COMMAND=$3
NODE_NAME=$(oc get pods -n $NAMESPACE  $POD_NAME --no-headers=true -ocustom-columns=node:.spec.nodeName)
CONTAINER_ID=$(oc get pods  -ojsonpath={.status.containerStatuses[0].containerID} -n $NAMESPACE $POD_NAME | awk -F "//" '{print $2}')
CONTAINER_PID=$(oc debug node/$NODE_NAME -- chroot /host crictl inspect --output go-template --template '{{.info.pid}}' $CONTAINER_ID 2>/dev/null)
oc debug node/$NODE_NAME -- nsenter nsenter -t $CONTAINER_PID -n $COMMAND

