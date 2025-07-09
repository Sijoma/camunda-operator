package mycustom

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
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/camunda/camunda-operator/api/v1alpha1"
	"github.com/camunda/camunda-operator/pkg/labels"
)

type Strategy struct{}

func (m Strategy) BuildResources(osc v1alpha1.OrchestrationCluster) ([]client.Object, error) {
	svcAcc := createServiceAccount(osc)
	svc := createService(osc)
	sts := createCamundaStatefulSet(osc)

	resources := []client.Object{svcAcc, svc, sts}
	return resources, nil
}

func createService(camunda v1alpha1.OrchestrationCluster) *corev1.Service {
	return &corev1.Service{
		TypeMeta: metav1.TypeMeta{
			Kind:       "Service",
			APIVersion: "v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      buildNameWithCore(camunda),
			Namespace: camunda.Namespace,
			Labels:    labels.Create(&camunda),
		},
		Spec: corev1.ServiceSpec{
			ClusterIP:                "None",
			Selector:                 labels.CreateSelector(&camunda),
			Ports:                    createServicePorts(),
			PublishNotReadyAddresses: true,
		},
	}
}

func createCamundaStatefulSet(
	camunda v1alpha1.OrchestrationCluster,
) *appsv1.StatefulSet {
	return &appsv1.StatefulSet{
		TypeMeta: metav1.TypeMeta{
			Kind:       "StatefulSet",
			APIVersion: "apps/v1",
		},
		ObjectMeta: metav1.ObjectMeta{
			Name:      camunda.Name,
			Namespace: camunda.Namespace,
			Labels:    labels.Create(&camunda),
		},
		Spec: appsv1.StatefulSetSpec{
			ServiceName:         buildNameWithCore(camunda),
			PodManagementPolicy: appsv1.ParallelPodManagement,
			Replicas:            ptr.To(camunda.Spec.ClusterSize),
			Selector: &metav1.LabelSelector{
				MatchLabels: labels.CreateSelector(&camunda),
			},
			Template:             createPodTemplate(camunda),
			VolumeClaimTemplates: createVolumeClaimTemplates(),
		},
	}
}

func createPodTemplate(camunda v1alpha1.OrchestrationCluster) corev1.PodTemplateSpec {
	// TODO: Make it override the static predefined values
	fullEnv := camunda.Spec.Env
	fullEnv = append(fullEnv, env(camunda)...)

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels.Create(&camunda),
		},
		Spec: corev1.PodSpec{
			SecurityContext:    createPodSecurityContext(),
			ServiceAccountName: buildNameWithCore(camunda),
			Containers: []corev1.Container{
				{
					Name:            "camunda",
					Image:           "camunda/camunda:" + camunda.Spec.Version,
					ImagePullPolicy: corev1.PullIfNotPresent,
					Resources:       camunda.Spec.Resources,
					Ports:           createPorts(),
					LivenessProbe:   livenessProbe(),
					ReadinessProbe:  readinessProbe(),
					StartupProbe:    startupProbe(),
					SecurityContext: securityContext(),
					Env:             fullEnv,
					EnvFrom:         camunda.Spec.EnvFrom,
					VolumeMounts:    createVolumeMounts(),
				},
			},
			Volumes: createVolumes(),
		},
	}
}

func createVolumes() []corev1.Volume {
	return []corev1.Volume{
		{
			Name: "tmp",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
		{
			Name: "exporters",
			VolumeSource: corev1.VolumeSource{
				EmptyDir: &corev1.EmptyDirVolumeSource{},
			},
		},
	}
}

func createVolumeMounts() []corev1.VolumeMount {
	return []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: "/usr/local/zeebe/data",
		},
		{
			Name:      "exporters",
			MountPath: "/exporters",
		},
		{
			Name:      "tmp",
			MountPath: "/tmp",
		},
	}
}

func createVolumeClaimTemplates() []corev1.PersistentVolumeClaim {
	return []corev1.PersistentVolumeClaim{
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
	}
}

func securityContext() *corev1.SecurityContext {
	return &corev1.SecurityContext{
		AllowPrivilegeEscalation: ptr.To(false),
		Privileged:               ptr.To(false),
		ReadOnlyRootFilesystem:   ptr.To(true),
		RunAsNonRoot:             ptr.To(true),
		RunAsUser:                ptr.To(int64(1001)),
		SeccompProfile:           &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
	}
}

func createPodSecurityContext() *corev1.PodSecurityContext {
	return &corev1.PodSecurityContext{
		FSGroup:        ptr.To(int64(1001)),
		RunAsNonRoot:   ptr.To(true),
		SeccompProfile: &corev1.SeccompProfile{Type: corev1.SeccompProfileTypeRuntimeDefault},
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
			Value: "identity,operate,tasklist,broker,consolidated-auth",
		},
		{
			Name:  "CAMUNDA_SECURITY_AUTHORIZATIONS_ENABLED",
			Value: "true",
		},
		{
			Name:  "CAMUNDA_SECURITY_AUTHENTICATION_UNPROTECTEDAPI",
			Value: "false",
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

		e = append(e, camundaDatabaseElasticsearch(
			camunda.Spec.Database.HostName,
			camunda.Spec.Database.UserName,
			camunda.Spec.Database.Password,
		)...)
		e = append(e, appDatabase(
			"OPERATE",
			camunda.Spec.Database.HostName,
			camunda.Spec.Database.UserName,
			camunda.Spec.Database.Password,
		)...)
		e = append(e, appDatabase(
			"TASKLIST",
			camunda.Spec.Database.HostName,
			camunda.Spec.Database.UserName,
			camunda.Spec.Database.Password,
		)...)
		e = append(e, zeebeElasticsearch(
			camunda.Spec.Database.HostName,
			camunda.Spec.Database.UserName,
			camunda.Spec.Database.Password,
		)...)
	}

	return e
}

func getPodAddresses(camunda v1alpha1.OrchestrationCluster) string {
	podAddresses := make([]string, camunda.Spec.ClusterSize)
	svc := createService(camunda)

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

func buildNameWithCore(camunda v1alpha1.OrchestrationCluster) string {
	return camunda.Name + "-core"
}
