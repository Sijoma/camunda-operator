package controller

import (
	"context"
	"fmt"
	"net/url"

	"github.com/sijoma/camunda-go-sdk/management"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/controller-runtime/pkg/log"

	corev1alpha1 "github.com/camunda/camunda-operator/api/v1alpha1"
)

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
