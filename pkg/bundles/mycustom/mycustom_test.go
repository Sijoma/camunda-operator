package mycustom

import (
	"testing"

	"github.com/stretchr/testify/assert"
	corev1 "k8s.io/api/core/v1"
)

func TestMergeEnvVars(t *testing.T) {
	t.Run("both slices empty", func(t *testing.T) {
		first := []corev1.EnvVar{}
		second := []corev1.EnvVar{}

		result := mergeEnvVars(first, second)

		assert.Empty(t, result, "Result should be empty when both input slices are empty")
	})

	t.Run("first slice empty", func(t *testing.T) {
		first := []corev1.EnvVar{}
		second := []corev1.EnvVar{
			{Name: "VAR1", Value: "value1"},
			{Name: "VAR2", Value: "value2"},
		}

		result := mergeEnvVars(first, second)

		assert.Len(t, result, 2, "Result should contain all variables from second slice")
		assert.Equal(t, "VAR1", result[0].Name, "Variables should be sorted alphabetically")
		assert.Equal(t, "VAR2", result[1].Name, "Variables should be sorted alphabetically")
	})

	t.Run("second slice empty", func(t *testing.T) {
		first := []corev1.EnvVar{
			{Name: "VAR1", Value: "value1"},
			{Name: "VAR2", Value: "value2"},
		}
		second := []corev1.EnvVar{}

		result := mergeEnvVars(first, second)

		assert.Len(t, result, 2, "Result should contain all variables from first slice")
		assert.Equal(t, "VAR1", result[0].Name, "Variables should be sorted alphabetically")
		assert.Equal(t, "VAR2", result[1].Name, "Variables should be sorted alphabetically")
	})

	t.Run("non-overlapping variables", func(t *testing.T) {
		first := []corev1.EnvVar{
			{Name: "VAR1", Value: "value1"},
			{Name: "VAR2", Value: "value2"},
		}
		second := []corev1.EnvVar{
			{Name: "VAR3", Value: "value3"},
			{Name: "VAR4", Value: "value4"},
		}

		result := mergeEnvVars(first, second)

		assert.Len(t, result, 4, "Result should contain all variables from both slices")
		assert.Equal(t, "VAR1", result[0].Name)
		assert.Equal(t, "VAR2", result[1].Name)
		assert.Equal(t, "VAR3", result[2].Name)
		assert.Equal(t, "VAR4", result[3].Name)
	})

	t.Run("overlapping variables", func(t *testing.T) {
		first := []corev1.EnvVar{
			{Name: "VAR1", Value: "value1-first"},
			{Name: "VAR2", Value: "value2-first"},
		}
		second := []corev1.EnvVar{
			{Name: "VAR1", Value: "value1-second"},
			{Name: "VAR3", Value: "value3-second"},
		}

		result := mergeEnvVars(first, second)

		assert.Len(t, result, 3, "Result should contain unique variables from both slices")
		assert.Equal(t, "VAR1", result[0].Name)
		assert.Equal(t, "value1-first", result[0].Value, "Value from first slice should override second slice")
		assert.Equal(t, "VAR2", result[1].Name)
		assert.Equal(t, "VAR3", result[2].Name)
	})

	t.Run("verify sorting", func(t *testing.T) {
		first := []corev1.EnvVar{
			{Name: "Z", Value: "z-value"},
			{Name: "X", Value: "x-value"},
		}
		second := []corev1.EnvVar{
			{Name: "B", Value: "b-value"},
			{Name: "A", Value: "a-value"},
		}

		result := mergeEnvVars(first, second)

		assert.Len(t, result, 4)
		assert.Equal(t, "A", result[0].Name, "Variables should be sorted alphabetically")
		assert.Equal(t, "B", result[1].Name, "Variables should be sorted alphabetically")
		assert.Equal(t, "X", result[2].Name, "Variables should be sorted alphabetically")
		assert.Equal(t, "Z", result[3].Name, "Variables should be sorted alphabetically")
	})

	t.Run("with ValueFrom", func(t *testing.T) {
		valueFrom := &corev1.EnvVarSource{
			FieldRef: &corev1.ObjectFieldSelector{
				FieldPath: "metadata.name",
			},
		}

		first := []corev1.EnvVar{
			{Name: "VAR1", ValueFrom: valueFrom},
		}
		second := []corev1.EnvVar{
			{Name: "VAR2", Value: "value2"},
		}

		result := mergeEnvVars(first, second)

		assert.Len(t, result, 2)
		assert.Equal(t, "VAR1", result[0].Name)
		assert.Equal(t, valueFrom, result[0].ValueFrom, "ValueFrom should be preserved")
		assert.Equal(t, "VAR2", result[1].Name)
	})
}
