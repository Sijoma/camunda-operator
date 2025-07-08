package labels

import corev1alpha1 "github.com/camunda/camunda-operator/api/v1alpha1"

func Create(osc *corev1alpha1.OrchestrationCluster) map[string]string {
	return map[string]string{
		"app":                          "camunda-platform",
		"app.kubernetes.io/name":       "camunda-platform",
		"app.kubernetes.io/instance":   osc.Name,
		"app.kubernetes.io/component":  "core",
		"app.kubernetes.io/part-of":    "camunda-platform",
		"app.kubernetes.io/managed-by": "orchestrationcluster-controller",
		"app.kubernetes.io/version":    osc.Spec.Version,
	}
}

// CreateSelector creates the selector labels for resources
func CreateSelector(osc *corev1alpha1.OrchestrationCluster) map[string]string {
	return map[string]string{
		"app":                          "camunda-platform",
		"app.kubernetes.io/name":       "camunda-platform",
		"app.kubernetes.io/instance":   osc.Name,
		"app.kubernetes.io/part-of":    "camunda-platform",
		"app.kubernetes.io/managed-by": "orchestrationcluster-controller",
		"app.kubernetes.io/component":  "core",
	}
}
