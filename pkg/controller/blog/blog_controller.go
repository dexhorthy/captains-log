/*
Copyright 2019 Dexter Horthy.

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

package blog

import (
	"context"
	"reflect"

	bloggingv1alpha1 "github.com/dexhorthy/captains-log/pkg/apis/blogging/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
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

var log = logf.Log.WithName("blog-controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new Blog Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBlog{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("blog-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to Blog
	err = c.Watch(&source.Kind{Type: &bloggingv1alpha1.Blog{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	// TODO(user): Modify this to be the types you create
	// Uncomment watch a Deployment created by Blog - change this for objects you create
	err = c.Watch(&source.Kind{Type: &appsv1.Deployment{}}, &handler.EnqueueRequestForOwner{
		IsController: true,
		OwnerType:    &bloggingv1alpha1.Blog{},
	})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBlog{}

// ReconcileBlog reconciles a Blog object
type ReconcileBlog struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a Blog object and makes changes based on the state read
// and what is in the Blog.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=blogging.dexhorthy.com,resources=blogs,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=blogging.dexhorthy.com,resources=blogs/status,verbs=get;update;patch
func (r *ReconcileBlog) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	// Fetch the Blog instance
	instance := &bloggingv1alpha1.Blog{}
	err := r.Get(context.TODO(), request.NamespacedName, instance)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	deploymentName := instance.Name
	siteConfigMapName := instance.Name
	contentConfigMapName := instance.Name + "-content"

	siteCM, siteConfigHash, err := r.buildSiteConfigMap(instance, siteConfigMapName)
	if err != nil {
		return reconcile.Result{}, err
	}

	contentCM := r.buildContentConfigMap(instance, contentConfigMapName)
	deploy := r.buildDeployemnt(deploymentName, instance, siteConfigMapName, contentConfigMapName, siteConfigHash)
	svc := r.buildService(deploymentName, instance)

	if err := controllerutil.SetControllerReference(instance, deploy, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	if err := controllerutil.SetControllerReference(instance, svc, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	if err := controllerutil.SetControllerReference(instance, siteCM, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	if err := controllerutil.SetControllerReference(instance, contentCM, r.scheme); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.ensureDeployment(deploy); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.ensureService(svc); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.ensureConfigMap(siteCM, true); err != nil {
		return reconcile.Result{}, err
	}

	if err := r.ensureConfigMap(contentCM, false); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBlog) ensureDeployment(deploy *appsv1.Deployment) error {
	// Check if the Deployment already exists
	found := &appsv1.Deployment{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: deploy.Name, Namespace: deploy.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Create(context.TODO(), deploy)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// TODO we need to do some strategic merge patching here, for now hack it because we know which fields the BlogPost will be changing
	deploy.Spec.Template.Spec.Containers[0].Env = found.Spec.Template.Spec.Containers[0].Env
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(deploy.Spec, found.Spec) {
		found.Spec = deploy.Spec
		log.Info("Updating Deployment", "namespace", deploy.Namespace, "name", deploy.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileBlog) ensureService(svc *corev1.Service) error {
	// Check if the Service already exists
	found := &corev1.Service{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: svc.Name, Namespace: svc.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating Service", "namespace", svc.Namespace, "name", svc.Name)
		err = r.Create(context.TODO(), svc)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// this isn't modifiable in CRD, so assume its unchanged
	svc.Spec.ClusterIP = found.Spec.ClusterIP
	// Update the found object and write the result back if there are any changes
	if !reflect.DeepEqual(svc.Spec, found.Spec) {
		// copy into desired to avoid immutable warning
		svc.Spec.ClusterIP = found.Spec.ClusterIP
		// then copy into update target
		found.Spec = svc.Spec
		log.Info("Updating Service", "namespace", svc.Namespace, "name", svc.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileBlog) ensureConfigMap(cm *corev1.ConfigMap, update bool) error {
	// Check if the ConfigMap already exists
	found := &corev1.ConfigMap{}
	err := r.Get(context.TODO(), types.NamespacedName{Name: cm.Name, Namespace: cm.Namespace}, found)
	if err != nil && errors.IsNotFound(err) {
		log.Info("Creating ConfigMap", "namespace", cm.Namespace, "name", cm.Name)
		err = r.Create(context.TODO(), cm)
		if err != nil {
			return err
		}
	} else if err != nil {
		return err
	}

	// Update the found object and write the result back if there are any changes
	if update && !reflect.DeepEqual(cm.Data, found.Data) {
		found.Data = cm.Data
		log.Info("Updating ConfigMap", "namespace", cm.Namespace, "name", cm.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}
