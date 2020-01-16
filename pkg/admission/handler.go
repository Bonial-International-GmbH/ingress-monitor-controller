package admission

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"

	"k8s.io/api/admission/v1beta1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/klog"
)

const jsonContentType = "application/json"

// universalDeserializer can convert raw data into Go objects that satisfy runtime.Object.
var universalDeserializer = serializer.NewCodecFactory(runtime.NewScheme()).UniversalDeserializer()

// AdmitFunc is the signature of a function that handles admission requests.
type AdmitFunc func(*v1beta1.AdmissionRequest) (*v1beta1.AdmissionResponse, error)

// HandlerFunc wraps the AdmitFunc f in an http.HandlerFunc.
func HandlerFunc(f AdmitFunc) http.HandlerFunc {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handleAdmissionRequest(w, r, f)
	})
}

// handleAdmissionRequest handles the http portion of a request prior to
// handing to an admit function.
func handleAdmissionRequest(w http.ResponseWriter, r *http.Request, admit AdmitFunc) {
	bytes, err := doHandleAdmissionRequest(w, r, admit)
	if err != nil {
		bytes = []byte(err.Error())
	}

	if _, err = w.Write(bytes); err != nil {
		klog.Errorf("could not write response: %v", err)
	}
}

// doHandleAdmissionRequest handles the http portion of a request prior to
// handing to an admit function.
func doHandleAdmissionRequest(w http.ResponseWriter, r *http.Request, admit AdmitFunc) ([]byte, error) {
	if r.Method != http.MethodPost {
		w.WriteHeader(http.StatusMethodNotAllowed)
		return nil, fmt.Errorf(`invalid method %q, only POST requests are allowed`, r.Method)
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("could not read request body: %v", err)
	}

	if contentType := r.Header.Get("Content-Type"); contentType != jsonContentType {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("unsupported content type %q, expected %q", contentType, jsonContentType)
	}

	// The AdmissionReview that was sent to the webhook
	requestedAdmissionReview := v1beta1.AdmissionReview{}

	_, _, err = universalDeserializer.Decode(body, nil, &requestedAdmissionReview)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, fmt.Errorf("could not deserialize request: %v", err)
	} else if requestedAdmissionReview.Request == nil {
		w.WriteHeader(http.StatusBadRequest)
		return nil, errors.New("malformed admission review: request is nil")
	}

	klog.V(4).Infof("handling request: %s", requestedAdmissionReview.Request.UID)

	// The AdmissionReview that will be returned
	responseAdmissionReview := v1beta1.AdmissionReview{}

	responseAdmissionReview.Response, err = admit(requestedAdmissionReview.Request)
	if err != nil {
		responseAdmissionReview.Response = &v1beta1.AdmissionResponse{
			Result: &metav1.Status{
				Message: err.Error(),
			},
		}
	}

	// Ensure that the request UID is set in the response
	responseAdmissionReview.Response.UID = requestedAdmissionReview.Request.UID

	klog.V(5).Infof("sending response: %v", responseAdmissionReview.Response)

	bytes, err := json.Marshal(responseAdmissionReview)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		return nil, fmt.Errorf("failed to marshal admission response: %v", err)
	}

	return bytes, nil
}
