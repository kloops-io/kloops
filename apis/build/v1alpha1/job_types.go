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
	runtime "k8s.io/apimachinery/pkg/runtime"
)

// JobState specifies the current pipelne status
type JobState string

// Various job types.
const (
	// TriggeredState for pipelines that have been triggered
	TriggeredState JobState = "triggered"
	// PendingState pipeline is pending
	PendingState JobState = "pending"
	// RunningState pipeline is running
	RunningState JobState = "running"
	// SuccessState pipeline is successful
	SuccessState JobState = "success"
	// FailureState failed
	FailureState JobState = "failure"
	// AbortedState aborted
	AbortedState JobState = "aborted"
	// ErrorState means the job could not schedule (bad config, perhaps).
	ErrorState JobState = "error"
)

// Pull describes a pull request at a particular point in time.
type Pull struct {
	Number int    `json:"number"`
	Author string `json:"author"`
	SHA    string `json:"sha"`
	Title  string `json:"title,omitempty"`

	// Ref is git ref can be checked out for a change
	// for example,
	// github: pull/123/head
	// gerrit: refs/changes/00/123/1
	Ref string `json:"ref,omitempty"`
	// Link links to the pull request itself.
	Link string `json:"link,omitempty"`
	// CommitLink links to the commit identified by the SHA.
	CommitLink string `json:"commitLink,omitempty"`
	// AuthorLink links to the author of the pull request.
	AuthorLink string `json:"authorLink,omitempty"`
}

// Refs describes how the repo was constructed.
type Refs struct {
	// Org is something like kubernetes or k8s.io
	Org string `json:"org"`
	// Repo is something like test-infra
	Repo string `json:"repo"`
	// RepoLink links to the source for Repo.
	RepoLink string `json:"repoLink,omitempty"`
	BaseRef  string `json:"baseRef,omitempty"`
	BaseSHA  string `json:"baseSha,omitempty"`
	// BaseLink is a link to the commit identified by BaseSHA.
	BaseLink string `json:"baseLink,omitempty"`

	Pulls []Pull `json:"pulls,omitempty"`

	// PathAlias is the location under <root-dir>/src
	// where this repository is cloned. If this is not
	// set, <root-dir>/src/github.com/org/repo will be
	// used as the default.
	PathAlias string `json:"pathAlias,omitempty"`
	// CloneURI is the URI that is used to clone the
	// repository. If unset, will default to
	// `https://github.com/org/repo.git`.
	CloneURI string `json:"cloneUri,omitempty"`
	// SkipSubmodules determines if submodules should be
	// cloned when the job is run. Defaults to true.
	SkipSubmodules bool `json:"skipSubmodules,omitempty"`
	// CloneDepth is the depth of the clone that will be used.
	// A depth of zero will do a full clone.
	CloneDepth int `json:"cloneDepth,omitempty"`
}

// JobSpec defines the desired state of Job
type JobSpec struct {
	// Refs is the code under test, determined at runtime
	Refs *Refs `json:"refs,omitempty"`

	Resource *runtime.RawExtension `json:"resource,omitempty"`
}

// JobStatus defines the observed state of Job
type JobStatus struct {
	// State is the full state of the job
	State JobState `json:"state,omitempty"`
	// ReportURL is the link that will be used in the commit status.
	ReportURL string `json:"reportURL,omitempty"`
	// StartTime is when the job was created.
	StartTime metav1.Time `json:"startTime,omitempty"`
	// CompletionTime is when the job finished reconciling and entered a terminal state.
	CompletionTime *metav1.Time `json:"completionTime,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Job is the Schema for the jobs API
type Job struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   JobSpec   `json:"spec,omitempty"`
	Status JobStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// JobList contains a list of Job
type JobList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Job `json:"items"`
}

func init() {
	SchemeBuilder.Register(&Job{}, &JobList{})
}
