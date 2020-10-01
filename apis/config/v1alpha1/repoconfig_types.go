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

// AutoMerge defines auto merge configuration
type AutoMerge struct {
	// BatchSizeLimitMap is the batch size limit as the value.
	// Special values:
	//  0 => unlimited batch size
	// -1 => batch merging disabled :(
	BatchSizeLimit int `json:"batchSizeLimit"`
	// MergeType is the merge method to use when merging pull requests.
	// Valid options are squash, rebase, and merge.
	MergeType PullRequestMergeType `json:"mergeType"`
	// Labels are the labels required on pull requests for merging
	Labels []string `json:"labels"`
	// MissingLabels are the labels that must not be present on pull requests for merging
	MissingLabels []string `json:"missingLabels"`
	// ReviewApprovedRequired tells that review must be approved on pull requests for merging
	ReviewApprovedRequired bool `json:"reviewApprovedRequired"`
}

// GitHubRepo defines a GitHub repository
type GitHubRepo struct {
	// Owner is the repository owner name
	Owner string `json:"owner"`
	// Repo is the repository owner name
	Repo string `json:"repo"`
	// ServerURL is the GitHub server url
	ServerURL string `json:"server,omitempty"`
	// HmacToken is the secret used to validate webhooks
	HmacToken Secret `json:"hmacToken"`
	// Token is the token used to interact with the git repository
	Token Secret `json:"token"`
}

// GiteaRepo defines a Gitea repository
type GiteaRepo struct {
	// Owner is the repository owner name
	Owner string `json:"owner"`
	// Repo is the repository owner name
	Repo string `json:"repo"`
	// ServerURL is the GitHub server url
	ServerURL string `json:"server,omitempty"`
	// HmacToken is the secret used to validate webhooks
	HmacToken Secret `json:"hmacToken"`
	// Token is the token used to interact with the git repository
	Token Secret `json:"token"`
}

// RepoPluginConfig defines a PluginConfig (it can be a ref or an inline spec)
type RepoPluginConfig struct {
	Ref     string            `json:"ref,omitempty"`
	Spec    *PluginConfigSpec `json:"spec,omitempty"`
	Plugins []string          `json:"plugins,omitempty"`
}

// RepoConfigSpec defines the desired state of RepoConfig
type RepoConfigSpec struct {
	// BotName is the bot name used by plugins
	BotName string `json:"botName,omitempty"`
	// GitHub defines the GitHub repository details
	GitHub *GitHubRepo `json:"gitHub,omitempty"`
	// Gitea defines the Gitea repository details
	Gitea *GiteaRepo `json:"gitea,omitempty"`
	// AutoMerge configuration for the repository
	AutoMerge *AutoMerge `json:"autoMerge,omitempty"`
	// PluginConfig defines the plugin configuration for the repository
	PluginConfig RepoPluginConfig `json:"pluginConfig"`
}

// RepoConfigStatus defines the observed state of RepoConfig
type RepoConfigStatus struct {
}

// +kubebuilder:object:root=true
// +kubebuilder:printcolumn:name=age,JSONPath=.metadata.creationTimestamp,type=date
// +kubebuilder:subresource:status

// RepoConfig is the Schema for the repoconfigs API
type RepoConfig struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              RepoConfigSpec   `json:"spec,omitempty"`
	Status            RepoConfigStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// RepoConfigList contains a list of RepoConfig
type RepoConfigList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []RepoConfig `json:"items"`
}

func init() {
	SchemeBuilder.Register(&RepoConfig{}, &RepoConfigList{})
}
