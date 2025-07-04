package bundles

import (
	"fmt"

	"github.com/Masterminds/semver/v3"
	"sigs.k8s.io/controller-runtime/pkg/client"

	"github.com/camunda/camunda-operator/api/v1alpha1"
	"github.com/camunda/camunda-operator/pkg/bundles/mycustom"
)

type VersionStrategy interface {
	BuildResources(v1alpha1.OrchestrationCluster) ([]client.Object, error)
}

type Bundle struct {
	core     v1alpha1.OrchestrationCluster
	strategy VersionStrategy
}

func (b Bundle) Resources() ([]client.Object, error) {
	if b.strategy == nil {
		return nil, fmt.Errorf("no strategy passed to Resources")
	}
	return b.strategy.BuildResources(b.core)

}

func New(osc v1alpha1.OrchestrationCluster) (*Bundle, error) {
	// Our current strategies
	strategies := map[string]VersionStrategy{">= 8.7.0-alpha1": mycustom.Strategy{}}

	return newWithStrategies(osc, strategies)
}

func newWithStrategies(osc v1alpha1.OrchestrationCluster, supportedStrategies map[string]VersionStrategy) (*Bundle, error) {
	version, err := semver.NewVersion(osc.Spec.Version)
	if err != nil {
		return nil, fmt.Errorf("invalid version format: %s", osc.Spec.Version)
	}
	for constraint, strategy := range supportedStrategies {
		constraintVersion, err := semver.NewConstraint(constraint)
		if err != nil {
			return nil, fmt.Errorf("invalid version constraint: %s", constraint)
		}
		if constraintVersion.Check(version) {
			return &Bundle{
				core:     osc,
				strategy: strategy,
			}, nil
		}
	}
	return nil, fmt.Errorf("no strategy found for version: %s", osc.Spec.Version)
}
