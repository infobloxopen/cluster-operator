package clustervalidator

import (
	"crypto/tls"
	"io/ioutil"
	"net/http"

	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/serializer"
	"k8s.io/apimachinery/pkg/util/json"
)

var (
	scheme          = runtime.NewScheme()
	codecs          = serializer.NewCodecFactory(scheme)
	tlscert, tlskey string
)
var log = logf.Log.WithName("validating_webhook")

type AdmissionController interface {
	HandleAdmission(review *v1beta1.AdmissionReview) error
}

type AdmissionControllerServer struct {
	AdmissionController AdmissionController
	Decoder             runtime.Decoder
}

func (acs *AdmissionControllerServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	body, err := ioutil.ReadAll(r.Body);
	if err != nil {
		log.Error(err, "Failed to read AdmissionReview request body.")
	}
    defer r.Body.Close()

	review := &v1beta1.AdmissionReview{}
	if _, _, err = acs.Decoder.Decode(body, nil, review); err != nil {
		log.Error(err, "Failed to decode AdmissionReview request.")
		return
	}

	acs.AdmissionController.HandleAdmission(review)
	responseInBytes, err := json.Marshal(review)
	if err != nil {
		log.Error(err, "Failed to write AdmissionReview response.")
		return
	}

	if _, err := w.Write(responseInBytes); err != nil {
		log.Error(err, "Failed to write AdmissionReview response.")
		return
	}
}

func GetAdmissionServerNoSSL(ac AdmissionController, listenOn string) *http.Server {
	server := &http.Server{
		Handler: &AdmissionControllerServer{
			AdmissionController: ac,
			Decoder:             codecs.UniversalDeserializer(),
		},
		Addr: listenOn,
	}

	return server
}

func GetAdmissionValidationServer(ac AdmissionController, tlsCert, tlsKey, listenOn string) (*http.Server, error) {
	sCert, err := tls.LoadX509KeyPair(tlsCert, tlsKey)
	if err != nil {
		return nil, err
	}
	server := GetAdmissionServerNoSSL(ac, listenOn)
	server.TLSConfig = &tls.Config{
		Certificates: []tls.Certificate{sCert},
	}
	return server, nil
}
