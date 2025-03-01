package specs

import (
	"fmt"
	"strconv"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/util/intstr"
	"k8s.io/utils/ptr"

	"github.com/camunda/camunda-operator/api/v1alpha1"
)

func CreateService(camunda v1alpha1.OrchestrationCluster) *corev1.Service {
	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      camunda.Name,
			Namespace: camunda.Namespace,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:                "None",
			Selector:                 createLabels(camunda),
			Ports:                    createServicePorts(),
			PublishNotReadyAddresses: true,
		},
	}
}

func createLabels(camunda v1alpha1.OrchestrationCluster) map[string]string {
	return map[string]string{
		"cluster":          camunda.Name,
		"operator-managed": "true",
	}
}

func CreateCamundaStatefulSet(
	camunda v1alpha1.OrchestrationCluster,
) *appsv1.StatefulSet {
	labels := createLabels(camunda)
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{
			Name:      camunda.Name,
			Namespace: camunda.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         camunda.Name,
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            ptr.To(camunda.Spec.ClusterSize),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels,
			},
			Template: createPodTemplate(camunda, labels),
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
						Resources: corev1.VolumeResourceRequirements{
							Requests: corev1.ResourceList{
								corev1.ResourceStorage: resource.MustParse("10Gi"),
							},
						},
					},
				},
			},
		},
	}
}

func createPodTemplate(camunda v1alpha1.OrchestrationCluster, labels map[string]string) corev1.PodTemplateSpec {
	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels,
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				{
					Name:           "camunda",
					Image:          "camunda/camunda:" + camunda.Spec.Version,
					Ports:          createPorts(),
					LivenessProbe:  livenessProbe(),
					ReadinessProbe: readinessProbe(),
					StartupProbe:   startupProbe(),
					Env:            env(camunda),
					VolumeMounts: []corev1.VolumeMount{
						{
							Name:      "data",
							MountPath: "/usr/local/zeebe/data",
						},
					},
				},
			},
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

func createServicePorts() []corev1.ServicePort {
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
				Port: intstr.FromString("management"),
				Path: "/actuator/health/readiness",
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
	}
}

func env(camunda v1alpha1.OrchestrationCluster) []corev1.EnvVar {
	e := []corev1.EnvVar{
		{
			Name: "ZEEBE_BROKER_CLUSTER_NODEID",
			ValueFrom: &corev1.EnvVarSource{
				FieldRef: &corev1.ObjectFieldSelector{
					APIVersion: "v1",
					FieldPath:  "metadata.labels['apps.kubernetes.io/pod-index']",
				},
			},
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_INITIALCONTACTPOINTS",
			Value: getPodAddresses(camunda),
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_PARTITIONS_COUNT",
			Value: strconv.Itoa(int(camunda.Spec.PartitionCount)),
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_REPLICATION_FACTOR",
			Value: strconv.Itoa(int(camunda.Spec.ReplicationFactor)),
		},
		{
			Name:  "ZEEBE_BROKER_CLUSTER_CLUSTER_SIZE",
			Value: strconv.Itoa(int(camunda.Spec.ClusterSize)),
		},
		{
			Name:  "SPRING_PROFILES_ACTIVE",
			Value: "broker,operate",
		},
	}

	if camunda.Spec.Database.Type == v1alpha1.ElasticsearchDatabaseType {
		e = append(e, camundaExporterEnv(
			camunda.Spec.Database.HostName,
			camunda.Spec.Database.UserName,
			camunda.Spec.Database.Password,
		)...)
		e = append(e, elasticsearchExporterEnv(
			camunda.Spec.Database.HostName,
			camunda.Spec.Database.UserName,
			camunda.Spec.Database.Password,
		)...)

		e = append(e, camundaDatabaseElasticsearch(camunda.Spec.Database.HostName, camunda.Spec.Database.UserName, camunda.Spec.Database.Password)...)
		e = append(e, operateDatabase(camunda.Spec.Database.HostName, camunda.Spec.Database.UserName, camunda.Spec.Database.Password)...)
		e = append(e, zeebeElasticsearch(camunda.Spec.Database.HostName, camunda.Spec.Database.UserName, camunda.Spec.Database.Password)...)
	}

	return e
}

func getPodAddresses(camunda v1alpha1.OrchestrationCluster) string {
	podAddresses := make([]string, camunda.Spec.ClusterSize)
	svc := CreateService(camunda)

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
