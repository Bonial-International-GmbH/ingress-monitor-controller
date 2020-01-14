package admission

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"k8s.io/api/admission/v1beta1"
	extensionsv1beta1 "k8s.io/api/extensions/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/klog"
)

const nginxWhitelistSourceRangeAnnotation = "nginx.ingress.kubernetes.io/whitelist-source-range"

var (
	// ingressResources are the GroupVersionResources that the webhook is
	// interested in. Resources that do not match any of the are ignored.
	ingressResources = []metav1.GroupVersionResource{
		{Group: "networking.k8s.io", Version: "v1beta1", Resource: "ingresses"},
		{Group: "extensions", Version: "v1beta1", Resource: "ingresses"},
	}
)

// patchOperation is an operation of a JSON patch as specified in by the RFC:
// https://tools.ietf.org/html/rfc6902.
type patchOperation struct {
	Op    string      `json:"op"`
	Path  string      `json:"path"`
	Value interface{} `json:"value,omitempty"`
}

// Webhook is a mutating admission webhook for ingress objects.
type Webhook struct {
	service monitor.Service
}

// NewWebhook creates a new admission webhook.
func NewWebhook(service monitor.Service) *Webhook {
	return &Webhook{
		service: service,
	}
}

// Admit handles an admission request and patches it if necessary. Patching may
// only happen on the nginx.ingress.kubernetes.io/whitelist-source-range
// annotation of an ingress object if necessary.
func (c *Webhook) Admit(ar *v1beta1.AdmissionRequest) (*v1beta1.AdmissionResponse, error) {
	if !isSupportedResource(ar.Resource) {
		klog.V(1).Infof("ignoring unsupported resource %v", ar.Resource)
		return allowedResponse(), nil
	}

	ingress := &extensionsv1beta1.Ingress{}

	_, _, err := universalDeserializer.Decode(ar.Object.Raw, nil, ingress)
	if err != nil {
		return nil, fmt.Errorf("failed to decode ingress object: %v", err)
	}

	klog.V(5).Infof("decoded ingress object %#v", ingress)

	if !shouldPatchSourceRangeWhitelist(ingress) {
		klog.V(4).Infof("ingress %s/%s does not require patching of source range whitelist", ingress.Namespace, ingress.Name)
		return allowedResponse(), nil
	}

	providerSourceRanges, err := c.service.GetProviderIPSourceRanges(ingress)
	if err != nil {
		klog.Errorf("skipping update of source range whitelist for ingress %s/%s due to: %v", ingress.Namespace, ingress.Name, err)
		return allowedResponse(), nil
	}

	if len(providerSourceRanges) == 0 {
		klog.V(4).Infof("no provider source ranges available for ingress %s/%s", ingress.Namespace, ingress.Name)
		return allowedResponse(), nil
	}

	sourceRanges, updated := mergeProviderSourceRanges(ingress, providerSourceRanges)
	if !updated {
		klog.V(4).Infof("no source range update needed for ingress %s/%s", ingress.Namespace, ingress.Name)
		return allowedResponse(), nil
	}

	patch, err := createSourceRangeWhitelistPatch(nginxWhitelistSourceRangeAnnotation, sourceRanges)
	if err != nil {
		return nil, fmt.Errorf("failed to create JSON patch for ingress %s/%s: %v", ingress.Namespace, ingress.Name, err)
	}

	klog.Infof("patching ingress %s/%s: %s", ingress.Namespace, ingress.Name, string(patch))

	return patchResponse(patch), nil
}

// allowedResponse is just a convenience func to signal that an object should
// be admitted to the cluster.
func allowedResponse() *v1beta1.AdmissionResponse {
	return &v1beta1.AdmissionResponse{Allowed: true}
}

// patchResponse admits an object to the cluster and also includes a JSON match
// with the mutation operations that should be performed on the object.
func patchResponse(patch []byte) *v1beta1.AdmissionResponse {
	pt := v1beta1.PatchTypeJSONPatch

	return &v1beta1.AdmissionResponse{
		Allowed:   true,
		Patch:     patch,
		PatchType: &pt,
	}
}

// isSupportedResource returns true if gvr matches one of the supported
// GroupVersionResources.
func isSupportedResource(gvr metav1.GroupVersionResource) bool {
	for _, res := range ingressResources {
		if res == gvr {
			return true
		}
	}

	return false
}

// shouldPatchSourceRangeWhitelist returns true if the source range whitelist
// of an ingress should be patched. Patching is necessary if the ingress has a
// monitor enabled and has configured the
// nginx.ingress.kubernetes.io/whitelist-source-range annotation to only allow
// traffic from whitelisted sources.
func shouldPatchSourceRangeWhitelist(ingress *extensionsv1beta1.Ingress) bool {
	annotations := config.Annotations(ingress.Annotations)

	if !annotations.BoolValue(config.AnnotationEnabled) {
		return false
	}

	return len(ingress.Annotations[nginxWhitelistSourceRangeAnnotation]) > 0
}

// createSourceRangeWhitelistPatch creates a JSON patch for the
// nginx.ingress.kubernetes.io/whitelist-source-range annotation of an ingress,
// which will replace it with the provided sourceRanges. It returns the
// marshaled bytes of the JSON patch and an error which is non-nil if
// marshaling the patch to JSON failed.
func createSourceRangeWhitelistPatch(annotationKey string, sourceRanges []string) ([]byte, error) {
	// Slashes need to be escaped with ~1 in json patches, see:
	// https://tools.ietf.org/html/rfc6901#section-3
	annotationKey = strings.ReplaceAll(annotationKey, "/", "~1")

	patchOperations := []patchOperation{
		{
			Op:    "replace",
			Path:  "/metadata/annotations/" + annotationKey,
			Value: strings.Join(sourceRanges, ","),
		},
	}

	return json.Marshal(patchOperations)
}

// mergeProviderSourceRanges merges the providerSourceRanges into the source
// ranges that are configured in the ingresses' whitelist and returns the final
// whitelist as slice of strings. It ensures that IP ranges that are already
// present are not added again. The second return value denotes whether the
// source ranges changed (true) or not (false).
func mergeProviderSourceRanges(ingress *extensionsv1beta1.Ingress, providerSourceRanges []string) ([]string, bool) {
	sourceRanges := strings.Split(ingress.Annotations[nginxWhitelistSourceRangeAnnotation], ",")
	missingSourceRanges := difference(providerSourceRanges, sourceRanges)

	if len(missingSourceRanges) == 0 {
		return sourceRanges, false
	}

	klog.V(4).Infof("missing source ranges: %v", missingSourceRanges)

	sourceRanges = append(sourceRanges, missingSourceRanges...)

	return sourceRanges, true
}

// difference returns elements that are in a but not in b.
func difference(a, b []string) []string {
	seen := make(map[string]struct{}, len(b))

	for _, el := range b {
		seen[el] = struct{}{}
	}

	var diff []string

	for _, el := range a {
		if _, found := seen[el]; !found {
			diff = append(diff, el)
		}
	}

	return diff
}
