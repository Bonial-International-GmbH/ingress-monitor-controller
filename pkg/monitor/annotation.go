package monitor

import (
	"strings"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/config"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/klog"
)

const nginxWhitelistSourceRangeAnnotation = "nginx.ingress.kubernetes.io/whitelist-source-range"

// AnnotateIngress implements Service.
func (s *service) AnnotateIngress(ingress *v1beta1.Ingress) (bool, error) {
	if !shouldPatchSourceRangeWhitelist(ingress) {
		klog.V(4).Infof("ingress %s/%s does not require patching of source range whitelist", ingress.Namespace, ingress.Name)
		return false, nil
	}

	providerSourceRanges, err := s.GetProviderIPSourceRanges(ingress)
	if err != nil {
		return false, err
	}

	if len(providerSourceRanges) == 0 {
		klog.V(4).Infof("no provider source ranges available for ingress %s/%s", ingress.Namespace, ingress.Name)
		return false, nil
	}

	sourceRanges := strings.Split(ingress.Annotations[nginxWhitelistSourceRangeAnnotation], ",")

	sourceRanges, updated := mergeProviderSourceRanges(sourceRanges, providerSourceRanges)
	if !updated {
		klog.V(4).Infof("no source range update needed for ingress %s/%s", ingress.Namespace, ingress.Name)
		return false, nil
	}

	klog.Infof("patching ingress %s/%s", ingress.Namespace, ingress.Name)

	ingress.Annotations[nginxWhitelistSourceRangeAnnotation] = strings.Join(sourceRanges, ",")

	return true, nil
}

// shouldPatchSourceRangeWhitelist returns true if the source range whitelist
// of an ingress should be patched. Patching is necessary if the ingress has a
// monitor enabled and has configured the
// nginx.ingress.kubernetes.io/whitelist-source-range annotation to only allow
// traffic from whitelisted sources.
func shouldPatchSourceRangeWhitelist(ingress *v1beta1.Ingress) bool {
	annotations := config.Annotations(ingress.Annotations)

	if !annotations.BoolValue(config.AnnotationEnabled) {
		return false
	}

	return len(ingress.Annotations[nginxWhitelistSourceRangeAnnotation]) > 0
}

// mergeProviderSourceRanges merges the providerSourceRanges into the source
// ranges that are configured in the ingresses' whitelist and returns the final
// whitelist as slice of strings. It ensures that IP ranges that are already
// present are not added again. The second return value denotes whether the
// source ranges changed (true) or not (false).
func mergeProviderSourceRanges(sourceRanges, providerSourceRanges []string) ([]string, bool) {
	missingSourceRanges := difference(providerSourceRanges, sourceRanges)

	if len(missingSourceRanges) == 0 {
		return sourceRanges, false
	}

	klog.Infof("missing source ranges: %v", missingSourceRanges)

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
