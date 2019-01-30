/*
Copyright 2019 Dexter Horthy.

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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// BlogPostSpec defines the desired state of BlogPost
type BlogPostSpec struct {
	// Blog is the name of the parent blog this post belongs to
	Blog string `json:"blog,omitempty"`
	// Content is the post content, with frontmatter
	Content string `json:"content,omitempty"`
}

// BlogPostStatus defines the observed state of BlogPost
type BlogPostStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file
}

// +genclient
// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BlogPost is the Schema for the blogposts API
// +k8s:openapi-gen=true
type BlogPost struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   BlogPostSpec   `json:"spec,omitempty"`
	Status BlogPostStatus `json:"status,omitempty"`
}

// +k8s:deepcopy-gen:interfaces=k8s.io/apimachinery/pkg/runtime.Object

// BlogPostList contains a list of BlogPost
type BlogPostList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []BlogPost `json:"items"`
}

func init() {
	SchemeBuilder.Register(&BlogPost{}, &BlogPostList{})
}
