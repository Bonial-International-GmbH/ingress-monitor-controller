package admission

import (
	"bytes"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/api/admission/v1beta1"
)

func TestHandleAdmissionRequest(t *testing.T) {
	tests := []struct {
		name               string
		requestMethod      string
		requestHeaders     map[string]string
		admitFunc          func(*v1beta1.AdmissionRequest) (*v1beta1.AdmissionResponse, error)
		requestBody        []byte
		expectedStatusCode int
		expectedResponse   []byte
	}{
		{
			name:               "only accepts POST requests",
			requestMethod:      "GET",
			expectedResponse:   []byte(`invalid method "GET", only POST requests are allowed`),
			expectedStatusCode: http.StatusMethodNotAllowed,
		},
		{
			name:               "request must have json content-type",
			requestMethod:      "POST",
			requestHeaders:     map[string]string{"Content-Type": "application/xml"},
			expectedResponse:   []byte(`unsupported content type "application/xml", expected "application/json"`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "request must contain valid json",
			requestMethod:      "POST",
			requestHeaders:     map[string]string{"Content-Type": "application/json"},
			requestBody:        []byte(`invalid json`),
			expectedResponse:   []byte(`could not deserialize request: couldn't get version/kind; json parse error: json: cannot unmarshal string into Go value of type struct { APIVersion string "json:\"apiVersion,omitempty\""; Kind string "json:\"kind,omitempty\"" }`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:               "admission review must contain a request",
			requestMethod:      "POST",
			requestHeaders:     map[string]string{"Content-Type": "application/json"},
			requestBody:        serializeObject(t, &v1beta1.AdmissionReview{}),
			expectedResponse:   []byte(`malformed admission review: request is nil`),
			expectedStatusCode: http.StatusBadRequest,
		},
		{
			name:           "executes admitFunc with request",
			requestMethod:  "POST",
			requestHeaders: map[string]string{"Content-Type": "application/json"},
			requestBody: serializeObject(t, &v1beta1.AdmissionReview{
				Request: newAdmissionRequestBuilder(t).build(),
			}),
			admitFunc: func(*v1beta1.AdmissionRequest) (*v1beta1.AdmissionResponse, error) {
				return allowedResponse(), nil
			},
			expectedResponse:   []byte(`{"response":{"uid":"","allowed":true}}`),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:           "attaches request UID to response",
			requestMethod:  "POST",
			requestHeaders: map[string]string{"Content-Type": "application/json"},
			requestBody: serializeObject(t, &v1beta1.AdmissionReview{
				Request: newAdmissionRequestBuilder(t).withUID("1234567890").build(),
			}),
			admitFunc: func(*v1beta1.AdmissionRequest) (*v1beta1.AdmissionResponse, error) {
				return allowedResponse(), nil
			},
			expectedResponse:   []byte(`{"response":{"uid":"1234567890","allowed":true}}`),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:           "json patch is included in the response",
			requestMethod:  "POST",
			requestHeaders: map[string]string{"Content-Type": "application/json"},
			requestBody: serializeObject(t, &v1beta1.AdmissionReview{
				Request: newAdmissionRequestBuilder(t).withUID("1234567890").build(),
			}),
			admitFunc: func(*v1beta1.AdmissionRequest) (*v1beta1.AdmissionResponse, error) {
				return newAdmissionResponseBuilder(t).
					withJSONPatch([]patchOperation{
						{
							Op: "add", Path: "/metadata/labels/foo", Value: "bar",
						},
					}).
					build(), nil
			},
			expectedResponse:   []byte(`{"response":{"uid":"1234567890","allowed":true,"patch":"W3sib3AiOiJhZGQiLCJwYXRoIjoiL21ldGFkYXRhL2xhYmVscy9mb28iLCJ2YWx1ZSI6ImJhciJ9XQ==","patchType":"JSONPatch"}}`),
			expectedStatusCode: http.StatusOK,
		},
		{
			name:           "admission errors produce rejected responses with status message",
			requestMethod:  "POST",
			requestHeaders: map[string]string{"Content-Type": "application/json"},
			requestBody: serializeObject(t, &v1beta1.AdmissionReview{
				Request: newAdmissionRequestBuilder(t).withUID("1234567890").build(),
			}),
			admitFunc: func(*v1beta1.AdmissionRequest) (*v1beta1.AdmissionResponse, error) {
				return nil, errors.New("nope, just nope")
			},
			expectedResponse:   []byte(`{"response":{"uid":"1234567890","allowed":false,"status":{"metadata":{},"message":"nope, just nope"}}}`),
			expectedStatusCode: http.StatusOK,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			server := httptest.NewServer(HandlerFunc(test.admitFunc))
			defer server.Close()

			buf := bytes.NewBuffer(test.requestBody)

			req, err := http.NewRequest(test.requestMethod, server.URL, buf)
			require.NoError(t, err)

			for key, val := range test.requestHeaders {
				req.Header.Set(key, val)
			}

			resp, err := http.DefaultClient.Do(req)
			require.NoError(t, err)
			defer resp.Body.Close()

			responseBody, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)

			assert.Equal(t, test.expectedStatusCode, resp.StatusCode)
			assert.Equal(t, string(test.expectedResponse), string(responseBody))
		})
	}
}
