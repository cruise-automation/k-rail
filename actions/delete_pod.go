package actions

import (
	"strconv"
	"time"

	log "github.com/sirupsen/logrus"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// DeletePod deletes a Pod by name from a given namespace
func DeletePod(client *kubernetes.Clientset, namespace, name string) (err error) {
	return client.CoreV1().Pods(namespace).Delete(name, &metav1.DeleteOptions{})
}

// DeletePodsByAnnotationAfterDuration deletes pods with a given label if the time since the value's epoch timstamp
// is greater than the given duration
func DeletePodsByAnnotationAfterDuration(
	client *kubernetes.Clientset,
	annotationName string,
	duration time.Duration) {

	pods, err := client.CoreV1().Pods("").List(metav1.ListOptions{})
	if err != nil {
		log.WithError(err).Error("could not list pods")
	}

	for _, pod := range pods.Items {
		for name, value := range pod.ObjectMeta.Annotations {
			if name == annotationName {
				i, err := strconv.ParseInt(value, 10, 64)
				if err != nil {
					log.WithFields(log.Fields{
						"kind":      "Pod",
						"resource":  pod.Name,
						"namespace": pod.Namespace,
					}).Error("could parse epoch time")
				}
				timestamp := time.Unix(i, 0)
				if time.Since(timestamp) > duration {
					err = DeletePod(client, pod.Namespace, pod.Name)
					if err != nil {
						log.WithFields(log.Fields{
							"kind":      "Pod",
							"resource":  pod.Name,
							"namespace": pod.Namespace,
						}).Error("could parse epoch time")
					}
				}
			}
		}
	}

}
