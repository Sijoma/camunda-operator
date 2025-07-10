package mycustom

import (
	"fmt"
	"strings"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"

	"github.com/camunda/camunda-operator/api/v1alpha1"
	"github.com/camunda/camunda-operator/pkg/labels"
)

func createHeadlessService(camunda v1alpha1.OrchestrationCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildNameWithCore(camunda) + "-headless",
			Namespace: camunda.Namespace,
			Labels:    labels.Create(&camunda),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:                "None",
			Selector:                 labels.CreateSelector(&camunda),
			Ports:                    createHeadlessServicePorts(),
			PublishNotReadyAddresses: true,
		},
	}
}

func createHeadlessServicePorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name: "management",
			Port: 9600,
		},
		{
			Name: "command",
			Port: 26501,
		},
		{
			Name: "internal",
			Port: 26502,
		},
	}
}

func createGatewayService(camunda v1alpha1.OrchestrationCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildNameWithCore(camunda) + "-gateway",
			Namespace: camunda.Namespace,
			Labels:    labels.Create(&camunda),
		},
		Spec: corev1.ServiceSpec{
			PublishNotReadyAddresses: true,
			Type:                     corev1.ServiceTypeClusterIP,
			Selector:                 labels.CreateSelector(&camunda),
			Ports:                    createGatewayPorts(),
		},
	}
}

func createGatewayPorts() []corev1.ServicePort {
	return []corev1.ServicePort{
		{
			Name: "http",
			Port: 8080,
		},
		{
			Name: "management",
			Port: 9600,
		},
		{
			Name: "gateway",
			Port: 26500,
		},
	}
}

func createPorts() []corev1.ContainerPort {
	return []corev1.ContainerPort{
		{
			Name:          "http",
			ContainerPort: 8080,
		},
		{
			Name:          "management",
			ContainerPort: 9600,
		},
		{
			Name:          "gateway",
			ContainerPort: 26500,
		},
		{
			Name:          "command",
			ContainerPort: 26501,
		},
		{
			Name:          "internal",
			ContainerPort: 26502,
		},
	}
}

func getPodAddresses(camunda v1alpha1.OrchestrationCluster) string {
	podAddresses := make([]string, camunda.Spec.ClusterSize)
	svc := createHeadlessService(camunda)

	for podIndex := int32(0); podIndex < camunda.Spec.ClusterSize; podIndex++ {
		podAddresses[podIndex] = fmt.Sprintf(
			"%s-%d.%s.%s.svc.cluster.local:26502",
			camunda.Name,
			podIndex,
			svc.Name,
			camunda.Namespace,
		)
	}
	return strings.Join(podAddresses, ",")
}

func livenessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromString("management"),
				Path: "/actuator/health/liveness",
			},
		},
	}
}

func readinessProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port:   intstr.FromString("management"),
				Path:   "/actuator/health/readiness",
				Scheme: corev1.URISchemeHTTP,
			},
		},
	}
}

func startupProbe() *corev1.Probe {
	return &corev1.Probe{
		ProbeHandler: corev1.ProbeHandler{
			HTTPGet: &corev1.HTTPGetAction{
				Port: intstr.FromString("management"),
				Path: "/actuator/health/startup",
			},
		},
		InitialDelaySeconds: 20,
		FailureThreshold:    30, // allow more time for startup
	}
}
