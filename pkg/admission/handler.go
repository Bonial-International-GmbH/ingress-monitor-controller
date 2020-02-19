package admission

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/Bonial-International-GmbH/ingress-monitor-controller/pkg/monitor"
	"k8s.io/api/extensions/v1beta1"
	"k8s.io/klog"
	"sigs.k8s.io/controller-runtime/pkg/webhook/admission"
)

// IngressHandler provides a mutating webhook that updates ingress annotations
// if necessary.
type IngressHandler struct {
	decoder        *admission.Decoder
	monitorService monitor.Service
}

// NewIngressHandler creates a new *IngressHandler which uses monitorService to
// update ingress annotations.
func NewIngressHandler(monitorService monitor.Service) *IngressHandler {
	return &IngressHandler{
		monitorService: monitorService,
	}
}

// Handle handles admission requests for ingress objects and will mutate their
// annotations if required. Handle implements admission.Handler.
func (h *IngressHandler) Handle(ctx context.Context, req admission.Request) admission.Response {
	ingress := &v1beta1.Ingress{}

	err := h.decoder.Decode(req, ingress)
	if err != nil {
		return admission.Errored(http.StatusBadRequest, err)
	}

	klog.V(5).Infof("decoded ingress object %#v", ingress)

	updated, err := h.monitorService.AnnotateIngress(ingress)
	if err != nil {
		klog.Errorf("skipping update of source range whitelist for ingress %s/%s due to: %v", ingress.Namespace, ingress.Name, err)
		return admission.Allowed("")
	}

	if !updated {
		return admission.Allowed("")
	}

	buf, err := json.Marshal(ingress)
	if err != nil {
		return admission.Errored(http.StatusInternalServerError, err)
	}

	return admission.PatchResponseFromRaw(req.Object.Raw, buf)
}

// InjectDecoder implements admission.DecoderInjector.
func (h *IngressHandler) InjectDecoder(d *admission.Decoder) error {
	h.decoder = d
	return nil
}
