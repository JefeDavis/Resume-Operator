/*
Copyright 2022.

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
	"errors"

	"github.com/nukleros/operator-builder-tools/pkg/controller/workload"
	"github.com/nukleros/operator-builder-tools/pkg/status"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

var ErrUnableToConvertProfile = errors.New("unable to convert to Profile")

// EDIT THIS FILE!  THIS IS SCAFFOLDING FOR YOU TO OWN!
// NOTE: json tags are required.  Any new fields you add must have json tags for the fields to be serialized.

// ProfileSpec defines the desired state of Profile.
type ProfileSpec struct {
	// INSERT ADDITIONAL SPEC FIELDS - desired state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	// +kubebuilder:validation:Optional
	Profile ProfileSpecProfile `json:"profile,omitempty"`

	// +kubebuilder:validation:Optional
	Web ProfileSpecWeb `json:"web,omitempty"`

	// +kubebuilder:default="example.com"
	// +kubebuilder:validation:Optional
	// (Default: "example.com")
	BaseURL string `json:"baseURL,omitempty"`

	// +kubebuilder:default="John Doe - CV"
	// +kubebuilder:validation:Optional
	// (Default: "John Doe - CV")
	PageTitle string `json:"pageTitle,omitempty"`

	// +kubebuilder:default="1"
	// +kubebuilder:validation:Optional
	// (Default: "1")
	PageCount string `json:"pageCount,omitempty"`

	// +kubebuilder:validation:Optional
	Pdf ProfileSpecPdf `json:"pdf,omitempty"`

	// PodLabels are additional labels applied to the pod templates of the
	// resume and pdf-converter deployments.
	// +kubebuilder:validation:Optional
	PodLabels map[string]string `json:"podLabels,omitempty"`

	// PodAnnotations are additional annotations applied to the pod templates
	// of the resume and pdf-converter deployments.
	// +kubebuilder:validation:Optional
	PodAnnotations map[string]string `json:"podAnnotations,omitempty"`

	// SecurityContext is the container-level security context applied to the
	// resume container.
	// +kubebuilder:validation:Optional
	SecurityContext *corev1.SecurityContext `json:"securityContext,omitempty"`

	// PodSecurityContext is the pod-level security context applied to the
	// resume and pdf-converter pod templates.
	// +kubebuilder:validation:Optional
	PodSecurityContext *corev1.PodSecurityContext `json:"podSecurityContext,omitempty"`

	// Ingress configures an Ingress for the resume site. It is used when Route
	// is not set. When omitted, a default nginx Ingress is created.
	// +kubebuilder:validation:Optional
	Ingress *ProfileSpecIngress `json:"ingress,omitempty"`

	// Route configures a Gateway API route (e.g. HTTPRoute) for the resume
	// site as an alternative to Ingress. When set, it takes precedence over
	// Ingress and no Ingress is created.
	// +kubebuilder:validation:Optional
	Route *ProfileSpecRoute `json:"route,omitempty"`
}

// ProfileSpecIngress configures the Ingress resource for the resume site.
type ProfileSpecIngress struct {
	// ClassName sets spec.ingressClassName on the Ingress.
	// +kubebuilder:default="nginx"
	// +kubebuilder:validation:Optional
	// (Default: "nginx")
	ClassName string `json:"className,omitempty"`

	// Annotations are added to the Ingress metadata. Supply controller-specific
	// settings here, e.g. cert-manager.io/cluster-issuer: letsencrypt-staging.
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels are added to the Ingress metadata.
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`
}

// ProfileSpecRoute configures a Gateway API route (HTTPRoute by default) for
// the resume site.
type ProfileSpecRoute struct {
	// APIVersion of the route resource to create.
	// +kubebuilder:default="gateway.networking.k8s.io/v1"
	// +kubebuilder:validation:Optional
	// (Default: "gateway.networking.k8s.io/v1")
	APIVersion string `json:"apiVersion,omitempty"`

	// Kind of the route resource to create.
	// +kubebuilder:default="HTTPRoute"
	// +kubebuilder:validation:Optional
	// (Default: "HTTPRoute")
	Kind string `json:"kind,omitempty"`

	// ParentRefs reference the Gateway(s) the route attaches to.
	// +kubebuilder:validation:Optional
	ParentRefs []ProfileSpecRouteParentRef `json:"parentRefs,omitempty"`

	// Hostname the route matches. Defaults to spec.baseURL when empty.
	// +kubebuilder:validation:Optional
	Hostname string `json:"hostname,omitempty"`

	// Annotations are added to the route metadata.
	// +kubebuilder:validation:Optional
	Annotations map[string]string `json:"annotations,omitempty"`

	// Labels are added to the route metadata.
	// +kubebuilder:validation:Optional
	Labels map[string]string `json:"labels,omitempty"`
}

// ProfileSpecRouteParentRef references a Gateway that a route attaches to.
type ProfileSpecRouteParentRef struct {
	// Name of the referenced Gateway.
	// +kubebuilder:validation:Required
	Name string `json:"name"`

	// Namespace of the referenced Gateway. Defaults to the Profile namespace.
	// +kubebuilder:validation:Optional
	Namespace string `json:"namespace,omitempty"`

	// SectionName of the Gateway listener to attach to.
	// +kubebuilder:validation:Optional
	SectionName string `json:"sectionName,omitempty"`

	// Port of the Gateway listener to attach to.
	// +kubebuilder:validation:Optional
	Port int32 `json:"port,omitempty"`

	// Group of the referenced parent resource.
	// +kubebuilder:default="gateway.networking.k8s.io"
	// +kubebuilder:validation:Optional
	// (Default: "gateway.networking.k8s.io")
	Group string `json:"group,omitempty"`

	// Kind of the referenced parent resource.
	// +kubebuilder:default="Gateway"
	// +kubebuilder:validation:Optional
	// (Default: "Gateway")
	Kind string `json:"kind,omitempty"`
}

type ProfileSpecProfile struct {
	// +kubebuilder:default="John"
	// +kubebuilder:validation:Optional
	// (Default: "John")
	FirstName string `json:"firstName,omitempty"`

	// +kubebuilder:default="Doe"
	// +kubebuilder:validation:Optional
	// (Default: "Doe")
	LastName string `json:"lastName,omitempty"`

	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	// (Default: "")
	PhoneNumber string `json:"phoneNumber,omitempty"`

	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	// (Default: "")
	Email string `json:"email,omitempty"`

	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	// (Default: "")
	LinkedinURL string `json:"linkedinURL,omitempty"`

	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	// (Default: "")
	GithubURL string `json:"githubURL,omitempty"`

	// +kubebuilder:default="South Carolina"
	// +kubebuilder:validation:Optional
	// (Default: "South Carolina")
	Location string `json:"location,omitempty"`

	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	// (Default: "")
	Overview string `json:"overview,omitempty"`

	// +kubebuilder:default={}
	// +kubebuilder:validation:Optional
	// (Default: "")
	CoreCompetencies []string `json:"coreCompetencies,omitempty"`

	// +kubebuilder:default={}
	// +kubebuilder:validation:optional
	// (Default: "")
	Projects []string `json:"projects,omitempty"`

	// +kubebuilder:default={}
	// +kubebuilder:validation:Optional
	// (Default: "")
	Skills []ProfileSpecSkillFamily `json:"skills,omitempty"`
}

type ProfileSpecSkillFamily struct {
	Family string   `json:"family,omitempty"`
	Items  []string `json:"items,omitempty"`
}

type ProfileSpecWeb struct {
	// +kubebuilder:validation:Optional
	Image ProfileSpecWebImage `json:"image,omitempty"`
}

type ProfileSpecWebImage struct {
	// +kubebuilder:default="latest"
	// +kubebuilder:validation:Optional
	// (Default: "latest")
	Tag string `json:"tag,omitempty"`

	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	// (Default: "")
	Registry string `json:"registry,omitempty"`

	// +kubebuilder:default="jefedavis/resume"
	// +kubebuilder:validation:Optional
	// (Default: "jefedavis/resume")
	Name string `json:"name,omitempty"`

	// +kubebuilder:default="IfNotPresent"
	// +kubebuilder:validation:Optional
	// (Default: "IfNotPresent")
	PullPolicy string `json:"pullPolicy,omitempty"`
}

type ProfileSpecPdf struct {
	// +kubebuilder:validation:Optional
	Image ProfileSpecPdfImage `json:"image,omitempty"`
}

type ProfileSpecPdfImage struct {
	// +kubebuilder:default=""
	// +kubebuilder:validation:Optional
	// (Default: "")
	Registry string `json:"registry,omitempty"`

	// +kubebuilder:default="jefedavis/resume"
	// +kubebuilder:validation:Optional
	// (Default: "jefedavis/resume")
	Name string `json:"name,omitempty"`

	// +kubebuilder:default="latest"
	// +kubebuilder:validation:Optional
	// (Default: "latest")
	Tag string `json:"tag,omitempty"`

	// +kubebuilder:default="IfNotPresent"
	// +kubebuilder:validation:Optional
	// (Default: "IfNotPresent")
	PullPolicy string `json:"pullPolicy,omitempty"`
}

// ProfileStatus defines the observed state of Profile.
type ProfileStatus struct {
	// INSERT ADDITIONAL STATUS FIELD - define observed state of cluster
	// Important: Run "make" to regenerate code after modifying this file

	Created               bool                     `json:"created,omitempty"`
	DependenciesSatisfied bool                     `json:"dependenciesSatisfied,omitempty"`
	Conditions            []*status.PhaseCondition `json:"conditions,omitempty"`
	Resources             []*status.ChildResource  `json:"resources,omitempty"`
}

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// Profile is the Schema for the profiles API.
type Profile struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`
	Spec              ProfileSpec   `json:"spec,omitempty"`
	Status            ProfileStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// ProfileList contains a list of Profile.
type ProfileList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []Profile `json:"items"`
}

// interface methods

// GetReadyStatus returns the ready status for a component.
func (component *Profile) GetReadyStatus() bool {
	return component.Status.Created
}

// SetReadyStatus sets the ready status for a component.
func (component *Profile) SetReadyStatus(ready bool) {
	component.Status.Created = ready
}

// GetDependencyStatus returns the dependency status for a component.
func (component *Profile) GetDependencyStatus() bool {
	return component.Status.DependenciesSatisfied
}

// SetDependencyStatus sets the dependency status for a component.
func (component *Profile) SetDependencyStatus(dependencyStatus bool) {
	component.Status.DependenciesSatisfied = dependencyStatus
}

// GetPhaseConditions returns the phase conditions for a component.
func (component *Profile) GetPhaseConditions() []*status.PhaseCondition {
	return component.Status.Conditions
}

// SetPhaseCondition sets the phase conditions for a component.
func (component *Profile) SetPhaseCondition(condition *status.PhaseCondition) {
	for i, currentCondition := range component.GetPhaseConditions() {
		if currentCondition.Phase == condition.Phase {
			component.Status.Conditions[i] = condition

			return
		}
	}

	// phase not found, lets add it to the list.
	component.Status.Conditions = append(component.Status.Conditions, condition)
}

// GetResources returns the child resource status for a component.
func (component *Profile) GetChildResourceConditions() []*status.ChildResource {
	return component.Status.Resources
}

// SetResources sets the phase conditions for a component.
func (component *Profile) SetChildResourceCondition(resource *status.ChildResource) {
	for i, currentResource := range component.GetChildResourceConditions() {
		if currentResource.Group == resource.Group && currentResource.Version == resource.Version && currentResource.Kind == resource.Kind {
			if currentResource.Name == resource.Name && currentResource.Namespace == resource.Namespace {
				component.Status.Resources[i] = resource

				return
			}
		}
	}

	// phase not found, lets add it to the collection
	component.Status.Resources = append(component.Status.Resources, resource)
}

// GetDependencies returns the dependencies for a component.
func (*Profile) GetDependencies() []workload.Workload {
	return []workload.Workload{}
}

// GetComponentGVK returns a GVK object for the component.
func (*Profile) GetWorkloadGVK() schema.GroupVersionKind {
	return GroupVersion.WithKind("Profile")
}

func init() {
	SchemeBuilder.Register(&Profile{}, &ProfileList{})
}
