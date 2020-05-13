package clustervalidator

import (
	"encoding/json"

	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"k8s.io/api/admission/v1beta1"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
)

type ClusterAdmission struct {
}

func UnmarshalClusterObject(rawReview []byte) (clusteroperatorv1alpha1.Cluster, error) {
	var cluster clusteroperatorv1alpha1.Cluster
	err := json.Unmarshal(rawReview, &cluster)

	return cluster, err
}

func (*ClusterAdmission) HandleAdmission(review *v1beta1.AdmissionReview) error {
	// Only operate on Cluster Kind
	switch review.Request.Kind.Kind {
	case "Cluster":
		switch review.Request.Operation {
		case "CREATE":
			// Approves all CREATE requests currently
			// Can be extended to run validation cases on CREATE
			review.Response = &v1beta1.AdmissionResponse{Allowed: true}
			break
		case "UPDATE":
			// rewiew.Request.Object and review.Request.OldObject contain the newly applyed and current objects
			// Both are RawExtension type, .Raw provides []byte that need to be unmarshalled
			oldCluster, unmarshalOldErr := UnmarshalClusterObject(review.Request.OldObject.Raw)
			if unmarshalOldErr != nil {
				return unmarshalOldErr
			}

			newCluster, unmarshalNewErr := UnmarshalClusterObject(review.Request.Object.Raw)
			if unmarshalNewErr != nil {
				return unmarshalNewErr
			}

			// Reject UPDATE if Spec.Name is different between cluster objects
			// Currently only case a Cluster object is rejected,
			// can extend to check for additional cases
			ValidateClusterName(oldCluster, newCluster, review)

			break
		}
	}
	return nil
}

// Validate Cluster Spec.Name field on UPDATE
// Currently only checks for change
func ValidateClusterName(oldCluster clusteroperatorv1alpha1.Cluster, newCluster clusteroperatorv1alpha1.Cluster, review *v1beta1.AdmissionReview) {
	if newCluster.Spec.Name != oldCluster.Spec.Name {
		review.Response = &v1beta1.AdmissionResponse{
			Allowed: false,
			Result: &v1.Status{
				Message: "Update rejected, Cluster Spec.Name cannot be updated.",
			},
		}
	} else {
		review.Response = &v1beta1.AdmissionResponse{
			Allowed: true,
			Result: &v1.Status{
				Message: "Update approved.",
			},
		}
	}
}
