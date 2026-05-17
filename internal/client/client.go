package client

import (
	"context"
	"fmt"

	apiextensionsv1 "k8s.io/apiextensions-apiserver/pkg/apis/apiextensions/v1"
	apiextclient "k8s.io/apiextensions-apiserver/pkg/client/clientset/clientset"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/cli-runtime/pkg/genericclioptions"
)

// Loads a CustomResourceDefinition from the API server
func FetchCRD(ctx context.Context, flags *genericclioptions.ConfigFlags, name string) (*apiextensionsv1.CustomResourceDefinition, error) {
	if name == "" {
		return nil, fmt.Errorf("CRD name must not be empty")
	}
	cfg, err := flags.ToRESTConfig()
	if err != nil {
		return nil, fmt.Errorf("building REST config: %w", err)
	}
	cs, err := apiextclient.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("building apiextensions clientset: %w", err)
	}
	crd, err := cs.ApiextensionsV1().CustomResourceDefinitions().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}
	return crd, nil
}
