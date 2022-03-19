module github.com/jefedavis/resume-operator

go 1.15

require (
	github.com/go-logr/logr v0.4.0
	github.com/nukleros/operator-builder-tools v0.2.0
	github.com/onsi/ginkgo v1.16.4
	github.com/onsi/gomega v1.15.0
	github.com/spf13/cobra v1.2.1
	github.com/stretchr/testify v1.7.0
	gopkg.in/yaml.v2 v2.4.0
	k8s.io/api v0.22.2
	k8s.io/apimachinery v0.22.2
	k8s.io/client-go v0.22.2
	sigs.k8s.io/controller-runtime v0.10.2
	sigs.k8s.io/kubebuilder/v3 v3.2.0
	sigs.k8s.io/yaml v1.2.0
)
