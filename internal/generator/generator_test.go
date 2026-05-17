package generator

import (
	"strings"
	"testing"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// schemaProp is a small constructor that returns its argument by value, so we
// can use a literal-as-pointer pattern when building nested schemas.
func schemaProp(p apiextensionsv1.JSONSchemaProps) apiextensionsv1.JSONSchemaProps {
	return p
}

func ptrBool(b bool) *bool { return &b }

func sampleCRD() *apiextensionsv1.CustomResourceDefinition {
	spec := apiextensionsv1.JSONSchemaProps{
		Type:        "object",
		Description: "WidgetSpec describes the desired state of a Widget.",
		Required:    []string{"size"},
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"size": {
				Type:        "string",
				Description: "Size selects the widget size.",
				Enum: []apiextensionsv1.JSON{
					{Raw: []byte(`"small"`)},
					{Raw: []byte(`"medium"`)},
					{Raw: []byte(`"large"`)},
				},
			},
			"replicas": {
				Type:        "integer",
				Description: "Replicas is the desired number of widget pods.",
			},
			"tags": {
				Type:        "array",
				Description: "Tags is a free-form list of labels.",
				Items: &apiextensionsv1.JSONSchemaPropsOrArray{
					Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"},
				},
			},
			"annotations": {
				Type:        "object",
				Description: "Annotations is a free-form map.",
				AdditionalProperties: &apiextensionsv1.JSONSchemaPropsOrBool{
					Schema: &apiextensionsv1.JSONSchemaProps{Type: "string"},
				},
			},
		},
	}

	root := apiextensionsv1.JSONSchemaProps{
		Type: "object",
		Properties: map[string]apiextensionsv1.JSONSchemaProps{
			"spec":   schemaProp(spec),
			"status": {Type: "object", Description: "should be skipped"},
		},
	}

	return &apiextensionsv1.CustomResourceDefinition{
		ObjectMeta: metav1.ObjectMeta{Name: "widgets.example.com"},
		Spec: apiextensionsv1.CustomResourceDefinitionSpec{
			Group: "example.com",
			Scope: apiextensionsv1.NamespaceScoped,
			Names: apiextensionsv1.CustomResourceDefinitionNames{
				Kind:     "Widget",
				Plural:   "widgets",
				Singular: "widget",
			},
			Versions: []apiextensionsv1.CustomResourceDefinitionVersion{
				{
					Name:    "v1alpha1",
					Served:  true,
					Storage: false,
					Schema:  &apiextensionsv1.CustomResourceValidation{OpenAPIV3Schema: &root},
				},
				{
					Name:    "v1",
					Served:  true,
					Storage: true,
					Schema:  &apiextensionsv1.CustomResourceValidation{OpenAPIV3Schema: &root},
				},
			},
		},
	}
}

func TestGenerate_BasicShape(t *testing.T) {
	out, err := Generate(sampleCRD())
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	got := string(out)

	mustContain(t, got, "apiVersion: example.com/v1")
	mustContain(t, got, "kind: Widget")
	mustContain(t, got, "name: example")
	mustContain(t, got, "namespace: default")
	mustContain(t, got, "spec:")

	// Required marker for size, optional for replicas.
	mustContain(t, got, "# required")
	mustContain(t, got, "# optional")

	// Enum surfaced in comment and first value used.
	mustContain(t, got, `# enum: ["small", "medium", "large"]`)
	mustContain(t, got, "size: small")

	// Status must not appear.
	if strings.Contains(got, "status:") {
		t.Errorf("status subresource leaked into output:\n%s", got)
	}

	// Storage version (v1) wins over v1alpha1.
	if strings.Contains(got, "v1alpha1") {
		t.Errorf("non-storage version leaked into output:\n%s", got)
	}
}

func TestGenerate_NoStorageVersion(t *testing.T) {
	crd := sampleCRD()
	for i := range crd.Spec.Versions {
		crd.Spec.Versions[i].Storage = false
	}
	if _, err := Generate(crd); err == nil {
		t.Fatal("expected error when no storage version is present")
	}
}

func TestGenerate_PreserveUnknownFields(t *testing.T) {
	crd := sampleCRD()
	v := storageVersionTestHelper(crd)
	v.Schema.OpenAPIV3Schema.Properties["spec"] = apiextensionsv1.JSONSchemaProps{
		Type:                   "object",
		XPreserveUnknownFields: ptrBool(true),
	}
	out, err := Generate(crd)
	if err != nil {
		t.Fatalf("Generate: %v", err)
	}
	if !strings.Contains(string(out), "spec:") {
		t.Errorf("expected spec key in output:\n%s", out)
	}
}

func storageVersionTestHelper(crd *apiextensionsv1.CustomResourceDefinition) *apiextensionsv1.CustomResourceDefinitionVersion {
	v, err := storageVersion(crd)
	if err != nil {
		panic(err)
	}
	return v
}

func mustContain(t *testing.T, haystack, needle string) {
	t.Helper()
	if !strings.Contains(haystack, needle) {
		t.Errorf("output missing %q\n--- output ---\n%s", needle, haystack)
	}
}
