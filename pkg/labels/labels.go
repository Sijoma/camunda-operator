package labels

import corev1alpha1 "github.com/camunda/camunda-operator/api/v1alpha1"

func Create(osc *corev1alpha1.OrchestrationCluster) map[string]string {
	l := commonLabels(osc)
	l["app.kubernetes.io/version"] = osc.Spec.Version
	return l
}

// CreateSelector creates the selector labels for resources.
// This should be used for labelSelector fields that are immutable, such as StatefulSets.
func CreateSelector(osc *corev1alpha1.OrchestrationCluster) map[string]string {
	return commonLabels(osc)
}

// commonLabels returns common labels for the OrchestrationCluster.
// The version is not included as it changes frequently and should not be used for a label selector for example.
func commonLabels(osc *corev1alpha1.OrchestrationCluster) map[string]string {
	return map[string]string{
		"app":                          "camunda-platform",
		"app.kubernetes.io/name":       "camunda-platform",
		"app.kubernetes.io/instance":   osc.Name,
		"app.kubernetes.io/component":  "core",
		"app.kubernetes.io/part-of":    "camunda-platform",
		"app.kubernetes.io/managed-by": "orchestrationcluster-controller",
	}
}
