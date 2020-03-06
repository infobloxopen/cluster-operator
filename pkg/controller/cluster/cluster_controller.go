package cluster

import (
	"context"
	"fmt"

	"github.com/seizadi/cluster-operator/kops"
	clusteroperatorv1alpha1 "github.com/seizadi/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"

	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_cluster")

// Add creates a new Cluster Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileCluster{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("cluster-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource Cluster
	err = c.Watch(&source.Kind{Type: &clusteroperatorv1alpha1.Cluster{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create that are owned by the primary resource
	// Watch for changes to secondary resource Pods and requeue the owner Cluster
	err = c.Watch(&source.Kind{Type: &corev1.Pod{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &clusteroperatorv1alpha1.Cluster{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileCluster implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileCluster{}

// ReconcileCluster reconciles a Cluster object
type ReconcileCluster struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Cluster object and makes changes based on the state read
// and what is in the Cluster.Spec
//
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileCluster) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling Cluster")
	// Fetch the Cluster instance
	instance := &clusteroperatorv1alpha1.Cluster{}
	fmt.Print("modify get \n")
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)

	//Finalizer name
	clusterFinalizer := "cluster.finalizer.go"
	fmt.Print("\n\n")
	fmt.Print(instance.Status)
	fmt.Print("\n\n")
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//The cluster is not waiting for deletion, so handle it normally
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		fmt.Print("im in the normal \n")
		// If no phase set default to pending for the initial phase
		if instance.Status.Phase == "" {
			fmt.Print("modify 1 \n")
			instance.Status.Phase = clusteroperatorv1alpha1.ClusterPending
		}

		// Add the finalizer and update the object
		if !contains(instance.ObjectMeta.Finalizers, clusterFinalizer) {
			fmt.Print("modify 2 \n")
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, clusterFinalizer)
		}

		// Run State Machine
		// PENDING -> SETUP -> DONE
		switch instance.Status.Phase {
		case clusteroperatorv1alpha1.ClusterPending:
			reqLogger.Info("Phase: PENDING")
			_, err := kops.CreateCluster(GetKopsConfig(instance.Spec.Name))
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Cluster Created")
			fmt.Print("modify 3 \n")
			instance.Status.Phase = clusteroperatorv1alpha1.ClusterSetup
		case clusteroperatorv1alpha1.ClusterSetup:
			reqLogger.Info("Phase: SETUP")
			status, err := kops.ValidateCluster(GetKopsConfig(instance.Spec.Name))
			if err != nil {
				return reconcile.Result{}, err
			}
			fmt.Print("modify 4 \n")
			instance.Status.KopsStatus = clusteroperatorv1alpha1.KopsStatus{}
			if len(status.Failures) > 0 {
				fmt.Print("modify 5 \n")
				instance.Status.KopsStatus.Failures = status.Failures
				reqLogger.Info("Cluster Not Ready")
			} else if len(status.Nodes) > 0 {
				fmt.Print("modify 6 \n")
				instance.Status.KopsStatus.Nodes = status.Nodes
				reqLogger.Info("Cluster Created")
				fmt.Print("modify 7 \n")
				instance.Status.Phase = clusteroperatorv1alpha1.ClusterDone
			} else {
				// FIXME - If we get this state try validate again!!!
				reqLogger.Info("Validate Returned Unexpected Result")
			}
		case clusteroperatorv1alpha1.ClusterDone:
			reqLogger.Info("Phase: DONE")
			return reconcile.Result{}, nil
		default:
			reqLogger.Info("NOP")
			return reconcile.Result{}, nil
		}

		// Update the Cluster instance, setting the status to the respective phase:
		fmt.Print("update 1 \n")
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}
		fmt.Print("update 2 \n")
		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}

		if instance.Status.Phase == clusteroperatorv1alpha1.ClusterSetup {
			// TODO - Set to time.Duration depending on back-off behavior
			//return reconcile.Result{RequeueAfter: time.Minute}, nil
			return reconcile.Result{Requeue: true}, nil
		}

		// Don't requeue. We should get called to reconcile when the CR changes.
		return reconcile.Result{}, nil

	} else if contains(instance.ObjectMeta.Finalizers, clusterFinalizer) {
		// our finalizer is present, so delete cluster first
		_, err := kops.DeleteCluster(GetKopsConfig(instance.Spec.Name))
		reqLogger.Info("Cluster Deleted")
		if err != nil {
			return reconcile.Result{}, err
		}

		// remove our finalizer from the list and update it.
		fmt.Print("modify 8 \n")
		instance.ObjectMeta.Finalizers = remove(instance.ObjectMeta.Finalizers, clusterFinalizer)
		fmt.Print("update 3 \n")
		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}

	}
	// Stop reconciliation as the item is being deleted
	return reconcile.Result{}, nil

	// Set Cluster instance as the owner and controller for AWS VPC Resource
	// if err := controllerutil.SetControllerReference(instance, vpc, r.scheme); err != nil {
	//	return reconcile.Result{}, err
	//}

}

// Get Kops Config Object
func GetKopsConfig(name string) clusteroperatorv1alpha1.KopsConfig {
	// Define a new Kops Cluster Config object
	return clusteroperatorv1alpha1.KopsConfig{
		Name:        name + ".soheil.belamaric.com",
		MasterCount: 1,
		MasterEc2:   "t2.micro",
		WorkerCount: 2,
		WorkerEc2:   "t2.micro",
		StateStore:  "s3://kops.state.seizadi.infoblox.com",
		Vpc:         "vpc-0a75b33895655b46a",
		Zones:       []string{"us-east-2a", "us-east-2b"},
	}
}

// Helper functions to check and remove string from a slice of strings.
func contains(slice []string, s string) bool {
	for _, item := range slice {
		if item == s {
			return true
		}
	}
	return false
}

func remove(slice []string, s string) (result []string) {
	for _, item := range slice {
		if item == s {
			continue
		}
		result = append(result, item)
	}
	return
}
