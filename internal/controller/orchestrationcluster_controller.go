/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/camunda/camunda-operator/api/v1alpha1"
	"github.com/camunda/camunda-operator/pkg/specs"
)

// OrchestrationClusterReconciler reconciles a OrchestrationCluster object
type OrchestrationClusterReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=core.camunda.io,resources=orchestrationclusters,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core.camunda.io,resources=orchestrationclusters/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=core.camunda.io,resources=orchestrationclusters/finalizers,verbs=update

// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete;deletecollection
// +kubebuilder:rbac:groups=core,resources=services/status,verbs=get

// CRUD apps: statefulsets
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets/scale,verbs=get;update
// +kubebuilder:rbac:groups=apps,resources=statefulsets/status,verbs=get

// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.2/pkg/reconcile
func (r *OrchestrationClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)

	orchestrationCluster := new(corev1alpha1.OrchestrationCluster)
	err := r.Client.Get(ctx, req.NamespacedName, orchestrationCluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	svc := specs.CreateService(*orchestrationCluster)
	newSvc := svc.Spec.DeepCopy()
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, svc, func() error {
		svc.Spec = *newSvc
		return ctrl.SetControllerReference(orchestrationCluster, svc, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	sts := specs.CreateCamundaStatefulSet(*orchestrationCluster)
	newSpec := sts.Spec.DeepCopy()
	var ready bool
	_, err = ctrl.CreateOrUpdate(ctx, r.Client, sts, func() error {
		sts.Spec = *newSpec
		// TODO: Check Topology for healthiness
		ready = sts.Status.ReadyReplicas > *newSpec.Replicas/2
		return ctrl.SetControllerReference(orchestrationCluster, sts, r.Scheme)
	})
	if err != nil {
		return ctrl.Result{}, err
	}

	// TODO: Implement proper status
	conditionStatus := metav1.ConditionFalse
	if ready {
		conditionStatus = metav1.ConditionTrue
	}
	changed := meta.SetStatusCondition(&orchestrationCluster.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             conditionStatus,
		ObservedGeneration: orchestrationCluster.Generation,
		Reason:             "CamundaReplicasReady",
		Message:            "replicas are ready",
	})
	if changed {
		err = r.Status().Update(ctx, orchestrationCluster)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrchestrationClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&corev1alpha1.OrchestrationCluster{}).
		Named("orchestrationcluster").
		Owns(&appsv1.StatefulSet{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
