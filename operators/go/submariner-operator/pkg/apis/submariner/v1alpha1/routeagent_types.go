package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// RouteagentSpec defines the desired state of Routeagent
// +k8s:openapi-gen=true
type RouteagentSpec struct {
	ServiceCIDR string `json:"serviceCIDR"`
	ClusterCIDR string `json:"clusterCIDR"`
	Debug       string `json:"debug"`
	ClusterID   string `json:"clusterID"`
	Namespace   string `json:"namespace"`
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// RouteagentStatus defines the observed state of Routeagent
// +k8s:openapi-gen=true
type RouteagentStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// Routeagent is the Schema for the routeagents API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=routeagents,scope=Namespaced
type Routeagent struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   RouteagentSpec   `json:"spec,omitempty"`
	Status RouteagentStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// RouteagentList contains a list of Routeagent
type RouteagentList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Routeagent `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Routeagent{}, &RouteagentList{})
}
