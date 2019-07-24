package v1alpha1

import meta_v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

// PodMonitor ...
type PodMonitor struct {
	meta_v1.TypeMeta   `json:",inline"`
	meta_v1.ObjectMeta `json:"metadata"`
	Spec               PodMonitorSpec   `json:"spec"`
	Status             PodMonitorStatus `json:"status,omitempty"`
}

// PodMonitorSpec ...
type PodMonitorSpec struct{}

// PodMonitorStatus ...
type PodMonitorStatus struct {
	PodPendingCount int32 `json:"podPendingCount,omitempty"`
	PodRunningCount int32 `json:"podRunningCount,omitempty"`
}

// PodMonitorList ...
type PodMonitorList struct {
	meta_v1.TypeMeta `json:",inline"`
	meta_v1.ListMeta `json:"metadata"`
	Items            []PodMonitor `json:"items"`
}
