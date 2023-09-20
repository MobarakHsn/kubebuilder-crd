package controller

import (
	"fmt"
	bookserverapi "github.com/MobarakHsn/kubebuilder-crd/api/v1"
	apps "k8s.io/api/apps/v1"
	core "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	meta "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

func (r *BookServerReconciler) EnsureDeployment() (ctrl.Result, error) {
	deployment := &apps.Deployment{}

	if err := r.Client.Get(r.ctx, types.NamespacedName{
		Namespace: r.bookServer.Namespace,
		Name:      r.bookServer.DeploymentName(),
	}, deployment); err != nil {
		if errors.IsNotFound(err) {
			fmt.Println("Could not find existing deployment for ", r.bookServer.Name, ", creating one...")
			deployment = r.NewDeployment()
			err := r.Client.Create(r.ctx, deployment)
			if err != nil {
				fmt.Printf("Error while creating deployment %s\n", err)
				return ctrl.Result{}, err
			} else {
				fmt.Printf("%s Deployments Created...\n", r.bookServer.Name)
			}
		} else {
			fmt.Printf("Error fetching deployment %s\n", err)
			return ctrl.Result{}, err
		}
	} else {
		if r.bookServer.Spec.Replicas != nil && *r.bookServer.Spec.Replicas != *deployment.Spec.Replicas {
			fmt.Println(*r.bookServer.Spec.Replicas, *deployment.Spec.Replicas)
			fmt.Println("Deployment replica miss match.....updating")
			deployment.Spec.Replicas = r.bookServer.Spec.Replicas
			if err := r.Client.Update(r.ctx, deployment); err != nil {
				fmt.Printf("Error updating deployment %s\n", err)
				return ctrl.Result{}, err
			}
			fmt.Println("Deployment updated")
		}
	}
	return ctrl.Result{}, nil
}

func (r *BookServerReconciler) NewDeployment() *apps.Deployment {
	labels := map[string]string{
		"app":  r.bookServer.Name,
		"kind": "BookServer",
	}
	return &apps.Deployment{
		TypeMeta: meta.TypeMeta{
			Kind: "Deployment",
		},
		ObjectMeta: meta.ObjectMeta{
			Name:      r.bookServer.DeploymentName(),
			Namespace: r.bookServer.Namespace,
			OwnerReferences: []meta.OwnerReference{
				*meta.NewControllerRef(r.bookServer, bookserverapi.GroupVersion.WithKind("BookServer")),
			},
		},
		Spec: apps.DeploymentSpec{
			Replicas: r.bookServer.Spec.Replicas,
			Selector: &meta.LabelSelector{
				MatchLabels: labels,
			},
			Template: core.PodTemplateSpec{
				ObjectMeta: meta.ObjectMeta{
					Namespace: r.bookServer.Namespace,
					Labels:    labels,
				},
				Spec: core.PodSpec{
					Containers: []core.Container{
						{
							Name:  r.bookServer.Name,
							Image: r.bookServer.Spec.Container.Image,
							Ports: []core.ContainerPort{
								{
									Name:          "http",
									ContainerPort: r.bookServer.Spec.Container.Port,
									Protocol:      "TCP",
								},
							},
						},
					},
				},
			},
		},
	}
}
