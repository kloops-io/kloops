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
	corev1 "k8s.io/api/core/v1"
)

// PullRequestMergeType inidicates the type of the pull request
type PullRequestMergeType string

// Possible types of merges for the GitHub merge API
const (
	MergeMerge  PullRequestMergeType = "merge"
	MergeRebase PullRequestMergeType = "rebase"
	MergeSquash PullRequestMergeType = "squash"
)

// IsValid checks that the merge type is valid
func (c PullRequestMergeType) IsValid() bool {
	return c == MergeMerge || c == MergeRebase || c == MergeSquash
}

// Secret defines a secret
type Secret struct {
	// Refers to a non-secret value
	Value string `json:"value,omitempty"`
	// Refers to a secret value to be used directly
	ValueFrom *ValueFrom `json:"valueFrom,omitempty"`
}

// ValueFrom defines a reference to a secret
type ValueFrom struct {
	SecretKeyRef corev1.SecretKeySelector `json:"secretKeyRef"`
}
