package mycustom

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/camunda/camunda-operator/api/v1alpha1"
	"github.com/camunda/camunda-operator/pkg/labels"
)

func createServiceAccount(camunda v1alpha1.OrchestrationCluster) *corev1.ServiceAccount {
	return &corev1.ServiceAccount{
		TypeMeta: metav1.TypeMeta{
			Kind:       "ServiceAccount",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildNameWithCore(camunda),
			Namespace: camunda.Namespace,
			Labels:    labels.Create(&camunda),
		},
	}
}
