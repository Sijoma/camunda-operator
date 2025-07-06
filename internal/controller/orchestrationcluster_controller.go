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
	"fmt"
	"net/url"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	k8sLabels "k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/sijoma/camunda-go-sdk/management"

	corev1alpha1 "github.com/camunda/camunda-operator/api/v1alpha1"
	"github.com/camunda/camunda-operator/pkg/bundles"
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

	bundle, err := bundles.New(*orchestrationCluster)
	if err != nil {
		log.FromContext(ctx).Error(err, "Error creating bundle for OrchestrationCluster",
			"cluster", orchestrationCluster.Name,
			"version", orchestrationCluster.Spec.Version,
		)
		return ctrl.Result{}, err
	}

	resources, err := bundle.Resources()
	if err != nil {
		log.FromContext(ctx).Error(err, "Error building resources for OrchestrationCluster",
			"cluster", orchestrationCluster.Name,
			"version", orchestrationCluster.Spec.Version,
		)
		return ctrl.Result{}, err
	}

	for _, resource := range resources {
		// Create or update the resource
		if err := ctrl.SetControllerReference(orchestrationCluster, resource, r.Scheme); err != nil {
			log.FromContext(ctx).Error(err, "Failed to set controller reference", "resource", resource.GetName())
			return ctrl.Result{}, err
		}

		merged := k8sLabels.Merge(resource.GetLabels(), managedLabels(orchestrationCluster))
		resource.SetLabels(merged)

		if err := r.Client.Patch(ctx, resource, client.Apply, client.ForceOwnership, client.FieldOwner("orchestrationcluster-controller")); err != nil {
			log.FromContext(ctx).Error(err, "Failed to create or patch resource", "resource", resource.GetName())
			return ctrl.Result{}, err
		}
	}

	err = r.checkCamunda(ctx, orchestrationCluster)
	if err != nil {
		log.FromContext(ctx).Error(err,
			"Error checking Camunda",
			"cluster", orchestrationCluster.Name,
		)
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

func (r *OrchestrationClusterReconciler) checkCamunda(ctx context.Context, cluster *corev1alpha1.OrchestrationCluster) error {
	actuatorPort := int32(9600)
	svc, err := lookupService(ctx, r.Client, cluster, actuatorPort)
	if err != nil {
		return fmt.Errorf("failed to lookup service for cluster %s: %w", cluster.Name, err)
	}

	actuatorURL := &url.URL{
		Scheme: "http",
		Host:   fmt.Sprintf("%s.%s.svc.cluster.local:%d", svc.Name, svc.Namespace, actuatorPort),
	}

	managementClient, err := management.NewClient(
		management.WithBaseURL(*actuatorURL),
	)
	if err != nil {
		return err
	}

	topo, err := managementClient.Cluster.Topology(ctx)
	if err != nil {
		return err
	}

	// Check if the cluster is ready
	ready := false

	if len(topo.PendingChange.Pending) > 0 {
		// If there are pending changes, we can assume that the cluster is scaling.
		// We can update the status or log this information as needed.
		log.FromContext(ctx).Info("Cluster is scaling", "pendingChanges", topo.PendingChange.Pending)
	} else {
		log.FromContext(ctx).Info("No pending changes in cluster topology")
	}

	if len(topo.Brokers) != int(cluster.Spec.ClusterSize) {
		log.FromContext(ctx).
			Info("Cluster size does not match desired size",
				"desiredSize", cluster.Spec.ClusterSize,
				"currentSize", len(topo.Brokers))
	}

	// TODO: Implement proper status
	if len(topo.Brokers) == int(cluster.Spec.ClusterSize) && topo.Version > 0 {
		ready = true
	}
	conditionStatus := metav1.ConditionTrue
	if ready {
		conditionStatus = metav1.ConditionTrue
	}
	changed := meta.SetStatusCondition(&cluster.Status.Conditions, metav1.Condition{
		Type:               "Ready",
		Status:             conditionStatus,
		ObservedGeneration: cluster.Generation,
		Reason:             "CamundaReplicasReady",
		Message:            "replicas are ready",
	})
	if changed {
		err = r.Status().Update(ctx, cluster)
		if err != nil {
			return err
		}
	}

	return nil
}

func managedLabels(cluster *corev1alpha1.OrchestrationCluster) map[string]string {
	return map[string]string{
		"app.kubernetes.io/managed-by": "orchestrationcluster-controller",
		"app.kubernetes.io/instance":   cluster.Name,
		"app.kubernetes.io/version":    cluster.Spec.Version,
	}
}

func lookupService(ctx context.Context, cli client.Client, cluster *corev1alpha1.OrchestrationCluster, desiredPort int32) (*corev1.Service, error) {
	selector := client.MatchingLabels(managedLabels(cluster))
	var svcList corev1.ServiceList
	if err := cli.List(ctx, &svcList, selector); err != nil {
		return nil, err
	}
	if len(svcList.Items) == 0 {
		return nil, fmt.Errorf("no service found for cluster %s", cluster.Name)
	}
	svc := &svcList.Items[0]
	var svcName string
	for _, port := range svc.Spec.Ports {
		if port.Port == desiredPort {
			svcName = svc.Name
			break
		}
	}
	if svcName == "" {
		return nil, fmt.Errorf("no service port 9600 found for cluster %s", cluster.Name)
	}

	return svc, nil
}
