package bundles

import (
	"testing"

	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/camunda/camunda-operator/api/v1alpha1"
)

type mockStrategy struct{}

func (m mockStrategy) BuildResources(osc v1alpha1.OrchestrationCluster) ([]client.Object, error) {
	return nil, nil
}

func TestNew(t *testing.T) {
	tests := []struct {
		name          string
		version       string
		registry      map[string]VersionStrategy
		expectError   bool
		errorContains string
	}{
		{
			name:        "Valid version with matching strategy",
			version:     "8.7.0",
			registry:    map[string]VersionStrategy{">= 8.7.0": mockStrategy{}},
			expectError: false,
		},
		{
			name:        "Valid version with matching strategy (higher version)",
			version:     "8.7.5",
			registry:    map[string]VersionStrategy{">= 8.7.0": mockStrategy{}},
			expectError: false,
		},
		{
			name:          "Valid version with no matching strategy",
			version:       "8.6.0",
			registry:      map[string]VersionStrategy{">= 8.7.0": mockStrategy{}},
			expectError:   true,
			errorContains: "no strategy found for version",
		},
		{
			name:          "Invalid version format",
			version:       "invalid",
			registry:      map[string]VersionStrategy{">= 8.7.0": mockStrategy{}},
			expectError:   true,
			errorContains: "invalid version format",
		},
		{
			name:          "Invalid version constraint in registry",
			version:       "8.7.0",
			registry:      map[string]VersionStrategy{"invalid": mockStrategy{}},
			expectError:   true,
			errorContains: "invalid version constraint",
		},
		{
			name:        "8.8.0-alpha5",
			version:     "8.8.0-alpha5",
			registry:    map[string]VersionStrategy{">= 8.7.0-alpha1": mockStrategy{}},
			expectError: false,
		},
		{
			name:        "8.7.0 matches also on alpha version pattern",
			version:     "8.7.0",
			registry:    map[string]VersionStrategy{">= 8.7.0-alpha1": mockStrategy{}},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test OrchestrationCluster with the specified version
			osc := v1alpha1.OrchestrationCluster{
				ObjectMeta: metav1.ObjectMeta{
					Name:      "test-cluster",
					Namespace: "default",
				},
				Spec: v1alpha1.OrchestrationClusterSpec{
					Version: tt.version,
				},
			}

			// Call the function being tested
			bundle, err := newWithStrategies(osc, tt.registry)

			// Check if we expect an error
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, bundle)
			} else {
				assert.NoError(t, err)
				assert.NotNil(t, bundle)
				assert.Equal(t, osc, bundle.core)
				assert.NotNil(t, bundle.strategy)
			}
		})
	}
}

// TestBundleBuildResources tests the Resources method of the Bundle struct
func TestBundleBuildResources(t *testing.T) {
	tests := []struct {
		name          string
		strategy      VersionStrategy
		expectError   bool
		errorContains string
	}{
		{
			name:          "Nil strategy",
			strategy:      nil,
			expectError:   true,
			errorContains: "no strategy passed to Resources",
		},
		{
			name:        "Valid strategy",
			strategy:    mockStrategy{},
			expectError: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a test Bundle with the specified strategy
			bundle := Bundle{
				core:     v1alpha1.OrchestrationCluster{},
				strategy: tt.strategy,
			}

			// Call the method being tested
			resources, err := bundle.Resources()

			// Check if we expect an error
			if tt.expectError {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errorContains)
				assert.Nil(t, resources)
			} else {
				assert.NoError(t, err)
				// We don't check the actual resources here since our mock returns nil
			}
		})
	}
}
