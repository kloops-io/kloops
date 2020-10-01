/*


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

// PluginConfigSpec defines the desired state of PluginConfig
type PluginConfigSpec struct {
	// Owners contains configuration related to handling OWNERS files.
	Owners Owners `json:"owners,omitempty"`
	// Built-in plugins specific configuration.
	Cat     Cat     `json:"cat,omitempty"`
	Goose   Goose   `json:"goose,omitempty"`
	Label   Label   `json:"label,omitempty"`
	Size    Size    `json:"size,omitempty"`
	Welcome Welcome `json:"welcome,omitempty"`
}

// Cat contains the configuration for the cat plugin.
type Cat struct {
	// Key is the api key for thecatapi.com
	Key Secret `json:"key"`
}

// Goose contains the configuration for the goose plugin.
type Goose struct {
	// Key is the api key for unsplash.com
	Key Secret `json:"key"`
}

// Label contains the configuration for the label plugin.
type Label struct {
	// AdditionalLabels is a set of additional labels enabled for use
	// on top of the existing "kind/*", "priority/*", and "area/*" labels.
	AdditionalLabels []string `json:"additionalLabels"`
}

// Size specifies configuration for the size plugin, defining lower bounds (in # lines changed) for each size label.
// XS is assumed to be zero.
type Size struct {
	S   int `json:"s"`
	M   int `json:"m"`
	L   int `json:"l"`
	Xl  int `json:"xl"`
	Xxl int `json:"xxl"`
}

// Welcome contains the configuration for the welcome plugin.
type Welcome struct {
	// MessageTemplate is the welcome message template to post on new-contributor PRs
	MessageTemplate string `json:"messageTemplate,omitempty"`
}

// Owners contains configuration related to handling OWNERS files.
type Owners struct {
	// MDYAMLRepos is a list of org and org/repo strings specifying the repos that support YAML
	// OWNERS config headers at the top of markdown (*.md) files. These headers function just like
	// the config in an OWNERS file, but only apply to the file itself instead of the entire
	// directory and all sub-directories.
	// The yaml header must be at the start of the file and be bracketed with "---" like so:
	/*
		---
		approvers:
		- mikedanese
		- thockin
		---
	*/
	MDYAMLRepos []string `json:"mdyamlrepos,omitempty"`
	// SkipCollaborators disables collaborator cross-checks and forces both
	// the approve and lgtm plugins to use solely OWNERS files for access
	// control in the provided repos.
	SkipCollaborators []string `json:"skipCollaborators,omitempty"`
	// LabelsExcludeList holds a list of labels that should not be present in any
	// OWNERS file, preventing their automatic addition by the owners-label plugin.
	// This check is performed by the verify-owners plugin.
	LabelsExcludeList []string `json:"labelsExcludes,omitempty"`
}

// PluginConfigStatus defines the observed state of PluginConfig
type PluginConfigStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name=age,JSONPath=.metadata.creationTimestamp,type=date
// +kubebuilder:subresource:status

// PluginConfig is the Schema for the pluginconfigs API
type PluginConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   PluginConfigSpec   `json:"spec,omitempty"`
	Status PluginConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// PluginConfigList contains a list of PluginConfig
type PluginConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []PluginConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&PluginConfig{}, &PluginConfigList{})
}
