package controller

import (
	"context"
	bookserverapi "github.com/MobarakHsn/kubebuilder-crd/api/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/fields"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

func trim(deploymentName string) string {
	if len(deploymentName) >= 11 {
		return deploymentName[:len(deploymentName)-11]
	}
	return deploymentName
}

// SetupWithManager sets up the controller with the Manager.
func (r *BookServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	if err := mgr.GetFieldIndexer().IndexField(context.Background(), &bookserverapi.BookServer{}, customDeployNameField, func(rawObj client.Object) []string {
		configDeployment := rawObj.(*bookserverapi.BookServer)
		if configDeployment.Spec.DeploymentName == "" {
			return nil
		}
		return []string{configDeployment.Spec.DeploymentName}
	}); err != nil {
		return err
	}

	handlerForDeployment := handler.EnqueueRequestsFromMapFunc(func(ctx context.Context, deployment client.Object) []reconcile.Request {
		//fmt.Println("There")
		attachedCustoms := &bookserverapi.BookServerList{}
		listOps := &client.ListOptions{
			FieldSelector: fields.OneTermEqualSelector(customDeployNameField, deployment.GetName()),
			Namespace:     deployment.GetNamespace(),
		}
		err := r.Client.List(context.TODO(), attachedCustoms, listOps)
		if err != nil {
			return []reconcile.Request{}
		}
		requests := make([]reconcile.Request, len(attachedCustoms.Items))
		for i, item := range attachedCustoms.Items {
			requests[i] = reconcile.Request{
				NamespacedName: types.NamespacedName{
					Name:      item.GetName(),
					Namespace: item.GetNamespace(),
				},
			}
		}
		return requests
	})
	return ctrl.NewControllerManagedBy(mgr).
		For(&bookserverapi.BookServer{}).
		Owns(&core.Service{}).
		Watches(
			&apps.Deployment{},
			handlerForDeployment,
		).
		Complete(r)
}

var (
	customDeployNameField = ".spec.deploymentName"
)
