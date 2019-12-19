package netservice

import (
	"context"
	"encoding/json"
	"net"

	svcctlv1alpha1 "operator-sdk/svcctl/pkg/apis/svcctl/v1alpha1"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("controller_netservice")

// Add creates a new NetService Controller and adds it to the Manager. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileNetService{client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("netservice-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to primary resource NetService
	err = c.Watch(&source.Kind{Type: &svcctlv1alpha1.NetService{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource services and requeue the owner NetService
	err = c.Watch(&source.Kind{Type: &corev1.Service{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &svcctlv1alpha1.NetService{},
	})
	if err != nil {
		return err
	}

	// Watch for changes to secondary resource services and requeue the owner NetService
	err = c.Watch(&source.Kind{Type: &corev1.Endpoints{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &svcctlv1alpha1.NetService{},
	})
	if err != nil {
		return err
	}

	return nil
}

// blank assignment to verify that ReconcileNetService implements reconcile.Reconciler
var _ reconcile.Reconciler = &ReconcileNetService{}

// ReconcileNetService reconciles a NetService object
type ReconcileNetService struct {
	// This client, initialized using mgr.Client() above, is a split client
	// that reads objects from the cache and writes to the apiserver
	client client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a NetService object and makes changes based on the state read
// and what is in the NetService.Spec
// Note:
// The Controller will requeue the Request to be processed again if the returned error is non-nil or
// Result.Requeue is true, otherwise upon completion it will remove the work from the queue.
func (r *ReconcileNetService) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)
	reqLogger.Info("Reconciling NetService")

	// Fetch the NetService instance
	cr := &svcctlv1alpha1.NetService{}
	err := r.client.Get(context.TODO(), request.NamespacedName, cr)
	if err != nil {
		if errors.IsNotFound(err) {
			// Request object not found, could have been deleted after reconcile request.
			// Return and don't requeue
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	templateName := cr.Name + "-template"

	// Create the template service, and set owner and controller
	tmpSvc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      templateName,
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			Selector:  cr.Spec.Selector,
			ClusterIP: "None",
		},
	}

	// Set CR instance as the owner and controller
	if err := controllerutil.SetControllerReference(cr, tmpSvc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check that the template service exists and is correct.
	svcFound := &corev1.Service{}
	name := types.NamespacedName{
		Name:      tmpSvc.Name,
		Namespace: tmpSvc.Namespace,
	}

	err = r.client.Get(context.TODO(), name, svcFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Template service does not exist - create it")
		err = r.client.Create(context.TODO(), tmpSvc)
	} else if err == nil {
		reqLogger.Info("We do not support updates to template service - assume OK")
		// This fails because we need to have the correct resourceVersion; not
		// copying for now. FIXME
		//err = r.client.Update(context.TODO(), tmpSvc)
	}

	if err != nil {
		reqLogger.Info("Error on template service")
		return reconcile.Result{}, err
	}

	// Create the real service, and set owner and controller
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Spec: corev1.ServiceSpec{
			ClusterIP: "None",
		},
	}

	if err := controllerutil.SetControllerReference(cr, svc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check that the template service exists and is correct.
	name = types.NamespacedName{
		Name:      svc.Name,
		Namespace: svc.Namespace,
	}

	err = r.client.Get(context.TODO(), name, svcFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Main service does not exist - create it")
		err = r.client.Create(context.TODO(), svc)
	} else if err == nil {
		reqLogger.Info("We do not support updates to main service - assume OK")
		// This fails because we need to have the correct resourceVersion; not
		// copying for now. FIXME
		//err = r.client.Update(context.TODO(), svc)
	}

	if err != nil {
		reqLogger.Info("Error on main service")
		return reconcile.Result{}, err
	}

	// Create the endpoints object, and set owner and controller
	ep := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      cr.Name,
			Namespace: cr.Namespace,
		},
		Subsets: []corev1.EndpointSubset{},
	}

	if err := controllerutil.SetControllerReference(cr, ep, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	// Check that the endpoints object exists
	epFound := &corev1.Endpoints{}
	name = types.NamespacedName{
		Name:      ep.Name,
		Namespace: ep.Namespace,
	}

	err = r.client.Get(context.TODO(), name, epFound)
	if err != nil && errors.IsNotFound(err) {
		reqLogger.Info("Endpoints object does not exist - create it")
		err = r.client.Create(context.TODO(), ep)
	} else if err == nil {
		reqLogger.Info("Endpoints object exists - just remember what we found")
		ep = epFound
	}

	if err != nil {
		reqLogger.Info("Error on endpoints object")
		return reconcile.Result{}, err
	}

	// Now find the relevant endpoints object for the template.
	tmpEpFound := &corev1.Endpoints{}
	name = types.NamespacedName{
		Name:      tmpSvc.Name,
		Namespace: tmpSvc.Namespace,
	}

	err = r.client.Get(context.TODO(), name, tmpEpFound)
	if err != nil {
		reqLogger.Info("Error reading templates endpoints object - timing window?")
		return reconcile.Result{}, err
	}

	if len(tmpEpFound.GetOwnerReferences()) == 0 {
		reqLogger.Info("Set owner references")
		err := controllerutil.SetControllerReference(cr, tmpEpFound, r.scheme)
		if err == nil {
			reqLogger.Info("Issue change to template endpoints")
			err = r.client.Update(context.TODO(), tmpEpFound)
		}
		if err != nil {
			reqLogger.Info("Error on templates endpoints object context")
			return reconcile.Result{}, err
		}
	}

	// One area for improvement - the net attach def here could be changed, and
	// we would never notice. Should probably put in an annotation in the
	// endpoints object, and check if it has changed. If so, discard all
	// addresses in endpoints and reload them. FIXME
	err = r.compareObjects(cr.Spec.NetAttachDef, request, tmpEpFound, ep)
	return reconcile.Result{}, err
}

// Compare the template and target objects.
func (r *ReconcileNetService) compareObjects(netName string, request reconcile.Request, template *corev1.Endpoints, target *corev1.Endpoints) error {
	reqLogger := log.WithValues("Request.Namespace", request.Namespace, "Request.Name", request.Name)

	changed := false

	targetAddresses := []corev1.EndpointAddress{}
	targetAddressMap := make(map[types.UID]corev1.EndpointAddress)

	if target.Subsets != nil && len(target.Subsets) != 0 {
		if target.Subsets[0].Addresses != nil {
			for _, address := range target.Subsets[0].Addresses {
				if address.TargetRef != nil {
					targetRef := address.TargetRef
					if targetRef.UID != "" {
						targetAddressMap[targetRef.UID] = address
					}
				}
			}
		}

		if len(target.Subsets) != 1 {
			// Press on regardless
			reqLogger.Info("Length of target subsets not 0 or 1 - should never happen")
		}
	}

	templateAddressMap := make(map[types.UID]corev1.EndpointAddress)

	if template.Subsets != nil && len(template.Subsets) != 0 {
		if template.Subsets[0].Addresses != nil {
			for _, address := range template.Subsets[0].Addresses {
				if address.TargetRef != nil {
					targetRef := address.TargetRef
					if targetRef.Kind == "Pod" && targetRef.UID != "" && targetRef.Name != "" && targetRef.Namespace != "" {
						templateAddressMap[targetRef.UID] = address
					}
				}
			}
		}

		if len(template.Subsets) != 1 {
			// Press on regardless
			reqLogger.Info("Length of template subsets not 0 or 1 - should never happen")
		}
	}

	for uid, address := range targetAddressMap {
		if _, ok := templateAddressMap[uid]; ok {
			// Address is already present, with correct values
			targetAddresses = append(targetAddresses, address)
		} else {
			// Address is in target, but not template. Rip it out.
			changed = true
			delete(targetAddressMap, uid)
			reqLogger.Info("Just removed address", "IP", address.IP)
		}
	}

	for uid, address := range templateAddressMap {
		if _, ok := targetAddressMap[uid]; !ok {
			// Address is in template, but not target. Add it.
			pod := &corev1.Pod{}
			name := types.NamespacedName{
				Name:      address.TargetRef.Name,
				Namespace: address.TargetRef.Namespace,
			}

			err := r.client.Get(context.TODO(), name, pod)
			if err != nil {
				// Keep calm and carry on
				reqLogger.Info("Failed to look up pod", "IP", address.IP, "Pod", address.TargetRef.Name, "err", err)
				continue
			}

			if pod.Annotations == nil {
				reqLogger.Info("No annotations in pod", "IP", address.IP, "Pod", address.TargetRef.Name)
				continue
			}

			networks, ok := pod.Annotations["k8s.v1.cni.cncf.io/networks-status"]
			if !ok {
				reqLogger.Info("No relevant annotations in pod", "IP", address.IP, "Pod", address.TargetRef.Name, "Annotations", pod.Annotations)
				continue
			}

			var results []map[string]interface{}
			json.Unmarshal([]byte(networks), &results)

			foundIP := ""

			for _, result := range results {
				// We panic on unexpected data. Could do better. FIXME
				resultNetName := result["name"].(string)
				ips := result["ips"].([]interface{})

				if resultNetName == netName {
					for _, ip := range ips {
						ipObject := net.ParseIP(ip.(string))

						if ipObject == nil {
							ipObject, _, _ = net.ParseCIDR(ip.(string))
						}

						if ipObject == nil || ipObject.To4() == nil {
							continue
						}

						foundIP = ipObject.String()
					}
				}
			}

			if foundIP == "" {
				reqLogger.Info("No IP in correct network found for pod", "IP", address.IP, "Pod", address.TargetRef.Name)
				continue
			}

			address.IP = foundIP
			targetAddresses = append(targetAddresses, address)
			changed = true
			reqLogger.Info("Just added address for pod", "IP", address.IP, "Pod", address.TargetRef.Name)
		}
	}

	reqLogger.Info("Completing", "Changed", changed)

	if changed {
		target.Subsets = []corev1.EndpointSubset{
			{
				Addresses: targetAddresses,
			},
		}

		err := r.client.Update(context.TODO(), target)
		if err != nil {
			reqLogger.Info("Oops")
			return err
		}
	}

	return nil
}
