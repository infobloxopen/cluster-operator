package cluster

import (
	"context"

	"github.com/infobloxopen/cluster-operator/kops"
	clusteroperatorv1alpha1 "github.com/infobloxopen/cluster-operator/pkg/apis/clusteroperator/v1alpha1"
	"github.com/infobloxopen/cluster-operator/utils"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"

	//metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	//"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
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
	err := r.client.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		reqLogger.Error(err, "error requesting instance")
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Owned objects are automatically garbage collected. For additional cleanup logic use finalizers.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	//Finalizer name
	clusterFinalizer := "cluster.finalizer.cluster-operator.infobloxopen.github.com"

	// TODO - We should maybe catch lack of kops configuration earlier in operator startup
	k, err := kops.NewKops()
	if err != nil {
		reqLogger.Error(err, "kops.NewKops Failed")
		return reconcile.Result{}, err
	}

	//If the cluster is not waiting for deletion, handle it normally
	if instance.ObjectMeta.DeletionTimestamp.IsZero() {
		// If no phase set default to pending for the initial phase
		if instance.Status.Phase == "" {
			instance.Status.Phase = clusteroperatorv1alpha1.ClusterPending
			config := instance.Spec.KopsConfig
			// If KopsConfig is not defined in CR, use default
			// TODO: Right now this only checks if the values are there. Eventually
			// we want to use a few inputs to pull information from the CMDB or
			// another controller that would hold the config information based on
			// the supplied infra info
			if config.Name == "" || config.MasterEc2 == "" || config.WorkerEc2 == "" || config.Vpc == "" ||
				config.StateStore == "" || config.MasterCount < 1 || config.WorkerCount < 1 || len(config.Zones) < 1 {
				instance.Spec.KopsConfig = GetKopsConfig(instance.Spec)
				if err := r.client.Update(context.TODO(), instance); err != nil {
					return reconcile.Result{}, err
				}
			}
		}

		// Add the finalizer and update the object
		if !utils.Contains(instance.ObjectMeta.Finalizers, clusterFinalizer) {
			instance.ObjectMeta.Finalizers = append(instance.ObjectMeta.Finalizers, clusterFinalizer)
			if err := r.client.Update(context.TODO(), instance); err != nil {
				return reconcile.Result{}, err
			}
		}

		// Run State Machine
		// PENDING -> SETUP -> DONE
		switch instance.Status.Phase {
		case clusteroperatorv1alpha1.ClusterPending:
			reqLogger.Info("Phase: PENDING")
			//cmd, err := k.CreateCluster(context.TODO(), GetKopsConfig(instance.Spec.Name))
			//if err != nil {
			//	reqLogger.Error(err, "error creating kops command")
			//	return reconcile.Result{}, err
			//}
			//if err := cmd.Start(); err != nil {
			//	reqLogger.Error(err, "error starting command")
			//	return reconcile.Result{}, err
			//}
			//if err := cmd.Wait(); err != nil {
			//	reqLogger.Error(err, "error waiting command")
			//	return reconcile.Result{}, err
			//}
			out, err := k.CreateCluster(GetKopsConfig(instance.Spec))
			reqLogger.Info(out)
			if err != nil {
				reqLogger.Error(err, "error creating kops command")
				return reconcile.Result{}, err
			}
			reqLogger.Info("Cluster Created")
			instance.Status.Phase = clusteroperatorv1alpha1.ClusterUpdate
		case clusteroperatorv1alpha1.ClusterUpdate:
			reqLogger.Info("Phase: UPDATE")
			out, err := k.UpdateCluster(GetKopsConfig(instance.Spec))
			reqLogger.Info(out)
			if err != nil {
				return reconcile.Result{}, err
			}
			// Some changes will require rebuilding the nodes (for example, resizing nodes or changing the AMI)
			// We call rolling-update to apply these changes
			out, err = k.RollingUpdateCluster(GetKopsConfig(instance.Spec))
			reqLogger.Info(out)
			if err != nil {
				return reconcile.Result{}, err
			}
			reqLogger.Info("Cluster Updated")
			instance.Status.Phase = clusteroperatorv1alpha1.ClusterSetup
		case clusteroperatorv1alpha1.ClusterSetup:
			reqLogger.Info("Phase: SETUP")
			status, err := k.ValidateCluster(instance.Spec.KopsConfig)
			if err != nil {
				return reconcile.Result{}, err
			}
			instance.Status.KopsStatus = clusteroperatorv1alpha1.KopsStatus{}
			if len(status.Failures) > 0 {
				instance.Status.KopsStatus.Failures = status.Failures
				reqLogger.Info("Cluster Not Ready")
			} else if len(status.Nodes) > 0 {
				instance.Status.KopsStatus.Nodes = status.Nodes
				reqLogger.Info("Cluster Created")
				instance.Status.Phase = clusteroperatorv1alpha1.ClusterDone
				config, err := k.GetKubeConfig(instance.Spec.KopsConfig)
				if err != nil {
					return reconcile.Result{}, err
				}
				instance.Status.KubeConfig = config
				reqLogger.Info("KUBECONFIG Updated")
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
		if err := r.client.Status().Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}

		if instance.Status.Phase == clusteroperatorv1alpha1.ClusterSetup {
			// TODO - Set to time.Duration depending on back-off behavior
			//return reconcile.Result{RequeueAfter: time.Minute}, nil
			return reconcile.Result{Requeue: true}, nil
		}

		// Don't requeue. We should get called to reconcile when the CR changes.
		return reconcile.Result{}, nil

	} else if utils.Contains(instance.ObjectMeta.Finalizers, clusterFinalizer) {
		// our finalizer is present, so delete cluster first
		out, err := k.DeleteCluster(instance.Spec.KopsConfig)
		reqLogger.Info(out)
		if err != nil {
			// FIXME - Ensure that delete implementation is idempotent and safe to invoke multiple times.
			// If we call delete and the cluster is not present it will cause error and it will keep erroring out
			//return reconcile.Result{}, err
		}

		reqLogger.Info("Cluster Deleted")
		// remove our finalizer from the list and update it.
		instance.ObjectMeta.Finalizers = utils.Remove(instance.ObjectMeta.Finalizers, clusterFinalizer)

		if err := r.client.Update(context.TODO(), instance); err != nil {
			return reconcile.Result{}, err
		}

	}
	// Stop reconciliation as the item is being deleted
	return reconcile.Result{}, nil
}

// Get Kops Config Object
func GetKopsConfig(c clusteroperatorv1alpha1.ClusterSpec) clusteroperatorv1alpha1.KopsConfig {

	// Define a new Kops Cluster Config object if not specified
	return clusteroperatorv1alpha1.KopsConfig{
		// FIXME - Pickup . .soheil.belamaric.com and "s3://kops.state.seizadi.infoblox.com" from Operator Config
		Name:        c.Name + ".soheil.belamaric.com",
		MasterCount: c.KopsConfig.MasterCount,
		MasterEc2:   c.KopsConfig.MasterEc2,
		WorkerCount: c.KopsConfig.WorkerCount,
		WorkerEc2:   c.KopsConfig.WorkerEc2,
		StateStore:  "s3://kops.state.seizadi.infoblox.com",
		Vpc:         c.KopsConfig.Vpc,
		Zones:       c.KopsConfig.Zones,
	}
}
