package mycustom

import (
	"flag"
	"testing"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/camunda/camunda-operator/api/v1alpha1"
	goldens "github.com/camunda/camunda-operator/pkg/golden"
)

var (
	update = flag.Bool("updategolden", true, "update the golden files of this test")
)

func apiSpec() v1alpha1.OrchestrationCluster {
	return v1alpha1.OrchestrationCluster{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "camunda-orchestration",
			Namespace: "camunda-orchestration-namespace",
		},
		Spec: v1alpha1.OrchestrationClusterSpec{
			Version:           "8.8.0-alpha1",
			PartitionCount:    3,
			ReplicationFactor: 3,
			ClusterSize:       3,
			Resources: corev1.ResourceRequirements{
				Requests: corev1.ResourceList{
					corev1.ResourceCPU:    resource.MustParse("100m"),
					corev1.ResourceMemory: resource.MustParse("100Mi"),
				},
			},
			EnvFrom: []corev1.EnvFromSource{{
				ConfigMapRef: &corev1.ConfigMapEnvSource{
					LocalObjectReference: corev1.LocalObjectReference{
						Name: "camunda-orchestration-configmap",
					},
				},
			}},
			Database: v1alpha1.Database{
				Type:     v1alpha1.ElasticsearchDatabaseType,
				UserName: "my-username",
				Password: corev1.SecretKeySelector{
					LocalObjectReference: corev1.LocalObjectReference{Name: "my-password-secret"},
				},
				HostName: "localhost:9205",
			},
		},
	}
}
func TestStatefulSetSpecs(t *testing.T) {

	got := createCamundaStatefulSet(apiSpec())
	golden, err := goldens.New(t, apiSpec().Name)
	if err != nil {
		t.Error("unable to create golden file", err)
	}

	err = golden.CheckOrUpdate(*update, got)
	if err != nil {
		t.Errorf("%s:\nerr:\n%v", apiSpec().Name, err)
	}
}

func TestServiceSpec(t *testing.T) {
	got := createService(apiSpec())
	golden, err := goldens.New(t, apiSpec().Name)
	if err != nil {
		t.Error("unable to create golden file", err)
	}

	err = golden.CheckOrUpdate(*update, got)
	if err != nil {
		t.Errorf("%s:\nerr:\n%v", apiSpec().Name, err)
	}
}
