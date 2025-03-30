package goldens_test

import (
	"flag"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	core "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	goldens "github.com/camunda/camunda-operator/pkg/golden"
)

var (
	//nolint:unused
	update = flag.Bool("updategolden", false, "update the golden files of this test")
)

var validPod = &core.Pod{
	ObjectMeta: metav1.ObjectMeta{
		Name: "goldentest",
	},
	Spec: core.PodSpec{
		Containers: []core.Container{
			{
				Name:  "busybox",
				Image: "busybox",
				Command: []string{
					"sleep",
					"1",
				},
			},
		},
	},
}

// TestGoldenFileWriteThenCheck tests CheckOrUpdate:
// 1. write a golden file when check is true
// 2. check the golden file when check is false
func TestGoldenFileWriteThenCheck(t *testing.T) {
	pod := validPod.DeepCopy()
	golden, err := goldens.New(t, "testWrite")
	require.NoError(t, err)
	assert.NotNil(t, golden)

	require.NoError(t, golden.CheckOrUpdate(true, pod))

	modifiedPod := pod.DeepCopy()
	modifiedPod.Name = uuid.NewString()

	require.NoError(t, golden.CheckOrUpdate(false, pod))
	require.Error(t, golden.CheckOrUpdate(false, modifiedPod))

}
