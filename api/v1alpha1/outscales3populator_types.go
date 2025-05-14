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

// OutscaleS3PopulatorSpec defines the desired state of OutscaleS3Populator
type OutscaleS3PopulatorSpec struct {
	// Bucket S3 d’origine
	Bucket string `json:"bucket"`

	// Objet (clé) dans le bucket
	Object string `json:"object"`

	// Endpoint S3 (ex: s3.eu-west-2.outscale.com)
	Endpoint string `json:"endpoint"`

	// Région
	Region string `json:"region"`

	// Credentials secrets Kubernetes
	AccessKeySecretRef corev1.SecretKeySelector `json:"accessKeySecretRef"`
	SecretKeySecretRef corev1.SecretKeySelector `json:"secretKeySecretRef"`
}

// OutscaleS3PopulatorStatus defines the observed state of OutscaleS3Populator
type OutscaleS3PopulatorStatus struct {
	Phase   string `json:"phase,omitempty"`
	Message string `json:"message,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status
// +kubebuilder:resource:path=outscales3populators,scope=Namespaced,shortName=oscp

// OutscaleS3Populator is the Schema for the outscales3populators API
type OutscaleS3Populator struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   OutscaleS3PopulatorSpec   `json:"spec,omitempty"`
	Status OutscaleS3PopulatorStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// OutscaleS3PopulatorList contains a list of OutscaleS3Populator
type OutscaleS3PopulatorList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []OutscaleS3Populator `json:"items"`
}

func init() {
	SchemeBuilder.Register(&OutscaleS3Populator{}, &OutscaleS3PopulatorList{})
}