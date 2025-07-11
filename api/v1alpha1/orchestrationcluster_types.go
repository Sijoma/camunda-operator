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

package v1alpha1

import (
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// OrchestrationClusterSpec defines the desired state of OrchestrationCluster.
type OrchestrationClusterSpec struct {
	// +default:value="8.7.7"
	Version           string `json:"version"`
	PartitionCount    int32  `json:"partitionCount,omitempty"`
	ReplicationFactor int32  `json:"replicationFactor,omitempty"`
	ClusterSize       int32  `json:"clusterSize,omitempty"`

	// Resources requirements of every generated Pod. Please refer to
	// https://kubernetes.io/docs/concepts/configuration/manage-resources-containers/
	// for more information.
	// +optional
	Resources corev1.ResourceRequirements `json:"resources,omitempty"`

	// Env to pass to the statefulset
	// +optional
	// +patchMergeKey=name
	// +patchStrategy=merge
	Env []corev1.EnvVar `json:"env,omitempty"`

	// EnvFrom to pass as source to the statefulset
	// +optional
	EnvFrom []corev1.EnvFromSource `json:"envFrom,omitempty"`

	Database Database `json:"database"`
}

type Database struct {
	// +kubebuilder:validation:Enum=elasticsearch;postgresql
	Type     DatabaseType             `json:"type"`
	UserName string                   `json:"userName,omitempty"`
	Password corev1.SecretKeySelector `json:"password,omitempty"`
	HostName string                   `json:"hostName,omitempty"`
}

type DatabaseType string

const ElasticsearchDatabaseType DatabaseType = "elasticsearch"
const PostgresqlDatabaseType DatabaseType = "postgresql"

// OrchestrationClusterStatus defines the observed state of OrchestrationCluster.
type OrchestrationClusterStatus struct {
	// +patchMergeKey=type
	// +patchStrategy=merge
	// +listType=map
	// +listMapKey=type
	Conditions []metav1.Condition `json:"conditions,omitempty" patchStrategy:"merge" patchMergeKey:"type" protobuf:"bytes,1,rep,name=conditions"` //nolint:lll
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// OrchestrationCluster is the Schema for the orchestrationclusters API.
type OrchestrationCluster struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OrchestrationClusterSpec   `json:"spec,omitempty"`
	Status OrchestrationClusterStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OrchestrationClusterList contains a list of OrchestrationCluster.
type OrchestrationClusterList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OrchestrationCluster `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OrchestrationCluster{}, &OrchestrationClusterList{})
}
