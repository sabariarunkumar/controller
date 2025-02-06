/*
Copyright 2025.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controller

import (
	"context"
	"fmt"
	managev1 "order-manager/api/v1"
	"regexp"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/tools/record"
	"k8s.io/utils/pointer"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/event"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/predicate"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

// OrderReconciler reconciles a Order object
type OrderReconciler struct {
	client.Client
	OrderToProcess chan managev1.Order
	CleanUpCtx     context.Context
	Recorder       record.EventRecorder
	Scheme         *runtime.Scheme
}

func (r *OrderReconciler) createSellerDeployment(order managev1.Order, namespace string) *appsv1.Deployment {
	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("seller-%s", order.Spec.Seller),
			Namespace: namespace,
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: pointer.Int32Ptr(1),
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{"seller": order.Spec.Seller},
			},
			// Dummy container configured currently. It should be implementing bussiness logic mentioned in function UpdateOrderStatusInSomeTimeMimicingSeller()
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: map[string]string{"seller": order.Spec.Seller}},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{
						{
							Name:  "order-processor",
							Image: "nginx:latest",
						},
					},
				},
			},
		},
	}
}
func (r *OrderReconciler) createSellerConfigMap(order managev1.Order, namespace string) *corev1.ConfigMap {
	return &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      fmt.Sprintf("seller-config-%s", order.Spec.Seller),
			Namespace: namespace,
			Labels: map[string]string{
				"process_orders": "true",
			},
			// OwnerReferences: []metav1.OwnerReference{
			// 	{
			// 		APIVersion: order.APIVersion,
			// 		Kind:       order.Kind,
			// 		Name:       order.Name,
			// 		UID:        order.UID,
			// 		Controller: pointer.Bool(true),
			// 	},
			// },
		},
		Data: map[string]string{
			order.Spec.ID: fmt.Sprintf("OrderName: %s, Product: %s, Qty: %d", order.Name, order.Spec.Inventory, order.Spec.Quantity),
		},
	}
}

// UpdateOrderStatusInSomeTimeMimicingSeller mimics seller application's implementation, where it process the order and update configmap
func (r *OrderReconciler) UpdateOrderStatusInSomeTimeMimicingSeller() {
	logger := log.FromContext(r.CleanUpCtx)
	var sellerOrders = make(map[string]map[string]string)
	t := time.NewTicker(30 * time.Second)
	for {
		select {
		case <-r.CleanUpCtx.Done():
			return
		case <-t.C:
			// logger.Info("Processing Orders at time ", "status_update", timeInterval.GoString())
			var sellersProcessed = []string{}
			for seller, orders := range sellerOrders {
				configMapName := fmt.Sprintf("seller-config-%s", seller)
				configMap := &corev1.ConfigMap{}
				// Tode: Namespace should be dynamic
				err := r.Get(r.CleanUpCtx, types.NamespacedName{Name: configMapName, Namespace: "default"}, configMap)
				if err != nil && apierrors.IsNotFound(err) {
					// Todo: define the custom logic
				} else {
					for orderID, orderDetails := range orders {
						configMap.Data[orderID] = orderDetails
						err = r.Update(r.CleanUpCtx, configMap)
						if err != nil {
							continue
							// Todo: define the custom logic|
						}

					}
				}
				sellersProcessed = append(sellersProcessed, seller)
				logger.Info("Procesed all the assinged Orders by [mimiced] seller ", "seller", seller)
			}
			for _, seller := range sellersProcessed {
				delete(sellerOrders, seller)
			}

		case currentOrder := <-r.OrderToProcess:
			seller := currentOrder.Spec.Seller
			if _, ok := sellerOrders[seller]; !ok {
				sellerOrders[seller] = make(map[string]string)
			}
			sellerOrders[seller][currentOrder.Spec.ID] = fmt.Sprintf("OrderName: %s, Product: %s, Qty: %d Processed", currentOrder.Name, currentOrder.Spec.Inventory, currentOrder.Spec.Quantity)
		}
	}

}

func getOrderNameIfProccessedBySeller(receivedOrder string) (order string, isProcessed bool) {
	re := regexp.MustCompile(`seller-processed-(.*)`) // Capture value after "OrderName:"
	match := re.FindStringSubmatch(receivedOrder)
	if len(match) > 1 {
		order = match[1]
		isProcessed = true
	}
	return
}

// +kubebuilder:rbac:groups=manage.sap,resources=orders,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=manage.sap,resources=orders/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=manage.sap,resources=orders/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Order object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.20.0/pkg/reconcile
func (r *OrderReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {

	logger := log.FromContext(ctx)

	// fmt.Println("request ", req)

	processedOrder, isProcessed := getOrderNameIfProccessedBySeller(req.Name)
	if isProcessed {
		var order managev1.Order
		if err := r.Get(ctx, types.NamespacedName{Name: processedOrder, Namespace: "default"}, &order); err != nil {
			return ctrl.Result{}, client.IgnoreNotFound(err)
		}
		// check if already processed
		if order.Status.State == "Processed" {
			return ctrl.Result{}, nil
		}
		order.Status.State = "Processed"
		if err := r.Status().Update(ctx, &order); err != nil {
			logger.Error(err, fmt.Sprintf("unable to update orderResource %s status", processedOrder))
			return ctrl.Result{}, err
		}

		r.Recorder.Event(&managev1.Order{}, "Normal", "SuccessfullyProcessed", fmt.Sprintf("Order Resource %s status has been recorded successfully", processedOrder))
		return ctrl.Result{}, nil
	}
	var order managev1.Order
	if err := r.Get(ctx, req.NamespacedName, &order); err != nil {
		logger.Error(err, fmt.Sprintf("unable to fetch orderResource %s ", req.Name))
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// logger.Info("Reconciling order resource", "name", order.Name)

	deployment := &appsv1.Deployment{}
	deploymentName := fmt.Sprintf("seller-%s", order.Spec.Seller)
	err := r.Get(ctx, types.NamespacedName{Name: deploymentName, Namespace: req.Namespace}, deployment)
	if err != nil && apierrors.IsNotFound(err) {
		logger.Info("Creating new Deployment for seller", "Seller", order.Spec.Seller)
		newDeployment := r.createSellerDeployment(order, req.Namespace)
		err = r.Create(ctx, newDeployment)
		if err != nil {
			return ctrl.Result{}, err
		}
	}

	configMapName := fmt.Sprintf("seller-config-%s", order.Spec.Seller)
	configMap := &corev1.ConfigMap{}
	err = r.Get(ctx, types.NamespacedName{Name: configMapName, Namespace: req.Namespace}, configMap)
	if err != nil && apierrors.IsNotFound(err) {
		logger.Info("Creating new ConfigMap for seller", "Seller", order.Spec.Seller)
		newConfigMap := r.createSellerConfigMap(order, req.Namespace)
		err = r.Create(ctx, newConfigMap)
		if err != nil {
			return ctrl.Result{}, err
		}
	} else {
		if strings.Contains(configMap.Data[order.Spec.ID], "Processed") {
			return ctrl.Result{}, nil
		}
		configMap.Data[order.Spec.ID] = fmt.Sprintf("OrderName: %s, Product: %s, Qty: %d", order.Name, order.Spec.Inventory, order.Spec.Quantity)
		err = r.Update(ctx, configMap)
		if err != nil {
			return ctrl.Result{}, err
		}

	}
	// push to mimic'ed seller for processing it

	r.OrderToProcess <- order
	logger.Info(fmt.Sprintf("Order Resource [%s] awaiting seller processing", order.Name))
	order.Status.State = "InProgress"
	if err := r.Status().Update(ctx, &order); err != nil {
		logger.Error(err, fmt.Sprintf("unable to update orderResource %s status", order.Name))
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *OrderReconciler) TriggerReconcileOnSellerProcessingOrder(ctx context.Context, obj client.Object) []reconcile.Request {
	cm := obj.(*corev1.ConfigMap)
	var requests []reconcile.Request

	for _, orderDetails := range cm.Data {
		re := regexp.MustCompile(`OrderName:\s*([^\s,]+)`) // Capture value after "OrderName:"

		match := re.FindStringSubmatch(orderDetails)
		if len(match) > 1 {
			if strings.Contains(orderDetails, "Processed") {
				requests = append(requests, reconcile.Request{
					NamespacedName: types.NamespacedName{
						Name:      "seller-processed-" + match[1],
						Namespace: cm.Namespace,
					},
				})
			}
		}
	}
	// fmt.Println("triggered requestes", len(requests))
	return requests
}

// SetupWithManager sets up the controller with the Manager.
func (r *OrderReconciler) SetupWithManager(mgr ctrl.Manager) error {

	return ctrl.NewControllerManagedBy(mgr).
		For(&managev1.Order{}).
		Watches(
			&corev1.ConfigMap{},
			handler.EnqueueRequestsFromMapFunc(r.TriggerReconcileOnSellerProcessingOrder),
		).
		WithEventFilter(predicate.Funcs{
			CreateFunc: func(e event.CreateEvent) bool {
				return e.Object.GetLabels()["process_orders"] == "true"
			},
			UpdateFunc: func(e event.UpdateEvent) bool {
				return e.ObjectNew.GetLabels()["process_orders"] == "true"
			},
		}).Complete(r)

}
