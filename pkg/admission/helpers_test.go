package admission

import (
	"bytes"
	gojson "encoding/json"
	"testing"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer/json"
	"k8s.io/apimachinery/pkg/types"
)

type admissionRequestBuilder struct {
	t   *testing.T
	req *v1beta1.AdmissionRequest
}

func newAdmissionRequestBuilder(t *testing.T) *admissionRequestBuilder {
	return &admissionRequestBuilder{
		t: t,
		req: &v1beta1.AdmissionRequest{
			Resource: metav1.GroupVersionResource{
				Group:    "networking.k8s.io",
				Version:  "v1beta1",
				Resource: "ingresses",
			},
		},
	}
}

func (ar *admissionRequestBuilder) withUID(uid string) *admissionRequestBuilder {
	ar.req.UID = types.UID(uid)
	return ar
}

func (ar *admissionRequestBuilder) withObject(obj runtime.Object) *admissionRequestBuilder {
	ar.req.Object = runtime.RawExtension{Raw: serializeObject(ar.t, obj)}
	return ar
}

func (ar *admissionRequestBuilder) withResource(gvr metav1.GroupVersionResource) *admissionRequestBuilder {
	ar.req.Resource = gvr
	return ar
}

func (ar *admissionRequestBuilder) build() *v1beta1.AdmissionRequest {
	return ar.req
}

type admissionResponseBuilder struct {
	t    *testing.T
	resp *v1beta1.AdmissionResponse
}

func newAdmissionResponseBuilder(t *testing.T) *admissionResponseBuilder {
	return &admissionResponseBuilder{
		t: t,
		resp: &v1beta1.AdmissionResponse{
			Allowed: true,
		},
	}
}

func (ar *admissionResponseBuilder) withJSONPatch(patch []patchOperation) *admissionResponseBuilder {
	buf, err := gojson.Marshal(patch)
	if err != nil {
		ar.t.Fatal(err)
	}

	pt := v1beta1.PatchTypeJSONPatch

	ar.resp.Patch = buf
	ar.resp.PatchType = &pt
	return ar
}

func (ar *admissionResponseBuilder) build() *v1beta1.AdmissionResponse {
	return ar.resp
}

func serializeObject(t *testing.T, obj runtime.Object) []byte {
	encoder := json.NewSerializer(json.DefaultMetaFactory, nil, nil, false)

	var buf bytes.Buffer

	err := encoder.Encode(obj, &buf)
	if err != nil {
		t.Fatal(err)
	}

	return buf.Bytes()
}
