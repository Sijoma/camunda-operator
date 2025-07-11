package mycustom

import (
	"sort"
	"strconv"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/utils/ptr"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/camunda/camunda-operator/api/v1alpha1"
	"github.com/camunda/camunda-operator/pkg/labels"
)

type Strategy struct{}

func (m Strategy) BuildResources(osc v1alpha1.OrchestrationCluster) ([]client.Object, error) {
	svcAcc := createServiceAccount(osc)
	headlessSvc := createHeadlessService(osc)
	gatewaySvc := createGatewayService(osc)
	sts := createCamundaStatefulSet(osc)

	resources := []client.Object{svcAcc, headlessSvc, gatewaySvc, sts}
	return resources, nil
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
			ServiceName:         createHeadlessService(camunda).Name,
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
	fullEnv := mergeEnvVars(env(camunda), camunda.Spec.Env)

	return corev1.PodTemplateSpec{
		ObjectMeta: metav1.ObjectMeta{
			Labels: labels.Create(&camunda),
		},
		Spec: corev1.PodSpec{
			SecurityContext:    createPodSecurityContext(),
			ServiceAccountName: createServiceAccount(camunda).Name,
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

func buildNameWithCore(camunda v1alpha1.OrchestrationCluster) string {
	return camunda.Name + "-core"
}

// mergeEnvVars returns the union of two []EnvVars, with any values set in first overriding those in second
func mergeEnvVars(first []corev1.EnvVar, second []corev1.EnvVar) []corev1.EnvVar {
	out := first
	if len(second) != 0 {
		existing := make(map[string]struct{}, len(first))

		for _, envVar := range first {
			existing[envVar.Name] = struct{}{}
		}

		for _, envVar := range second {
			if _, exists := existing[envVar.Name]; !exists {
				out = append(out, envVar)
			}
		}
	}

	// map order is undefined so sort for stable test output
	sort.Slice(out, func(i, j int) bool {
		return out[i].Name < out[j].Name
	})

	return out
}
