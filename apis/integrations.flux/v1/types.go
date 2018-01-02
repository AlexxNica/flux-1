package v1

import (
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"k8s.io/apimachinery/pkg/runtime"
)

// +genclient
// +genclient:noStatus
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FluxHelmResource represents custom resource associated with a Helm Chart
type FluxHelmResource struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec FluxHelmResourceSpec `json:"spec"`
}

// FluxHelmResourceSpec is the spec for a Foo resource
type FluxHelmResourceSpec struct {
	Image        string `json:"image"`
	ImageVersion string `json:"image-version,omitempty"`
	ImageTag     string `json:"image-tag,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// FluxHelmResourceList is a list of FluxHelmResource resources
type FluxHelmResourceList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata"`

	Items []FluxHelmResource `json:"items"`
}
