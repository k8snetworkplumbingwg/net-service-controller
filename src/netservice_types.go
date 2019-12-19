package v1alpha1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// NetServiceSpec defines the desired state of NetService
// +k8s:openapi-gen=true
type NetServiceSpec struct {
	// The name of the network attachment definition in question
	NetAttachDef string `json:"netAttachDef"`

	// Ports is not set, so no ports are possible. That's because we only
	// support headless services anyway, i.e. a DNS name, so it doesn't matter
	// except for informational purposes, and also because it ensures that we
	// don't have multiple EndpointSubsets within the same Endpoints object
	// anywhere, something that can happen with named targetPorts.

	// Selector for service.
	Selector map[string]string `json:"selector,omitempty" protobuf:"bytes,2,rep,name=selector"`
}

// NetServiceStatus defines the observed state of NetService
// +k8s:openapi-gen=true
type NetServiceStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "operator-sdk generate k8s" to regenerate code after modifying this file
	// Add custom validation using kubebuilder tags: https://book-v1.book.kubebuilder.io/beyond_basics/generating_crd.html
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetService is the Schema for the netservices API
// +k8s:openapi-gen=true
// +kubebuilder:subresource:status
type NetService struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   NetServiceSpec   `json:"spec,omitempty"`
	Status NetServiceStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// NetServiceList contains a list of NetService
type NetServiceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []NetService `json:"items"`
}

func init() {
	SchemeBuilder.Register(&NetService{}, &NetServiceList{})
}
