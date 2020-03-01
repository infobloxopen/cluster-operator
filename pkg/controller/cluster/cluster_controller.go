package cluster

import (
	"context"
	
	clusteroperatorv1alpha1 "github.com/seizadi/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"github.com/seizadi/cluster-operator/kops"
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

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

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
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
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
	
	// If no phase set default to pending for the initial phase
	if instance.Status.Phase == "" {
		instance.Status.Phase = clusteroperatorv1alpha1.ClusterPending
	}
	
	
	// Set Cluster instance as the owner and controller for AWS VPC Resource
	// if err := controllerutil.SetControllerReference(instance, vpc, r.scheme); err != nil {
	//	return reconcile.Result{}, err
	//}
	
	// Run State Machine
	// PENDING -> SETUP -> DONE
	switch instance.Status.Phase {
	case clusteroperatorv1alpha1.ClusterPending:
		reqLogger.Info("Phase: PENDING")
		result, err := kops.CreateCluster(GetKopsConfig(instance.Name))
		if err != nil {
			return reconcile.Result{}, err
		}
		reqLogger.Info("Cluster Created", result)
		instance.Status.Phase = clusteroperatorv1alpha1.ClusterSetup
	case clusteroperatorv1alpha1.ClusterSetup:
		reqLogger.Info("Phase: SETUP")
		status, err := kops.ValidateCluster(GetKopsConfig(instance.Name))
		if err != nil {
			return reconcile.Result{}, err
		}
		
		instance.Status.KopsStatus = clusteroperatorv1alpha1.KopsStatus {}
		if len(status.Failures) > 0 {
			instance.Status.KopsStatus.Failures = status.Failures
			reqLogger.Info("Cluster Not Ready")
		} else if len(status.Nodes) > 0 {
			instance.Status.KopsStatus.Nodes = status.Nodes
			reqLogger.Info("Cluster Created")
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
	
	// Update the At instance, setting the status to the respective phase:
	err = r.Status().Update(context.TODO(), instance)
	if err != nil {
		return reconcile.Result{}, err
	}
	
	if instance.Status.Phase == clusteroperatorv1alpha1.ClusterSetup {
		// TODO - Set to time.Duration depending on back-off behavior
		//return reconcile.Result{RequeueAfter: time.Minute}, nil
		return reconcile.Result{Requeue: true}, nil
	}
	
	// Don't requeue. We should get called to reconcile when the CR changes.
	return reconcile.Result{}, nil
}

// Get Kops Config Object
func GetKopsConfig(name string) clusteroperatorv1alpha1.KopsConfig {
	// Define a new Kops Cluster Config object
	return clusteroperatorv1alpha1.KopsConfig{
		Name: name + ".soheil.belamaric.com",
		MasterCount: 1,
		MasterEc2: "t2.micro",
		WorkerCount: 2,
		WorkerEc2: "t2.micro",
		StateStore: "s3://kops.state.seizadi.infoblox.com",
		Vpc: "vpc-0a75b33895655b46a",
		Zones: [] string {"us-east-2a", "us-east-2b"},
	}
}

