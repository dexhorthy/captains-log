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

package blogpost

import (
	"context"
	"crypto/sha1"
	"encoding/json"
	"fmt"
	"io"
	"reflect"

	bloggingv1alpha1 "github.com/dexhorthy/captains-log/pkg/apis/blogging/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller"
	"sigs.k8s.io/controller-runtime/pkg/handler"
	"sigs.k8s.io/controller-runtime/pkg/manager"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
	logf "sigs.k8s.io/controller-runtime/pkg/runtime/log"
	"sigs.k8s.io/controller-runtime/pkg/source"
)

var log = logf.Log.WithName("blogpost-controller")

/**
* USER ACTION REQUIRED: This is a scaffold file intended for the user to modify with their own Controller
* business logic.  Delete these comments after modifying this file.*
 */

// Add creates a new BlogPost Controller and adds it to the Manager with default RBAC. The Manager will set fields on the Controller
// and Start it when the Manager is Started.
func Add(mgr manager.Manager) error {
	return add(mgr, newReconciler(mgr))
}

// newReconciler returns a new reconcile.Reconciler
func newReconciler(mgr manager.Manager) reconcile.Reconciler {
	return &ReconcileBlogPost{Client: mgr.GetClient(), scheme: mgr.GetScheme()}
}

// add adds a new Controller to mgr with r as the reconcile.Reconciler
func add(mgr manager.Manager, r reconcile.Reconciler) error {
	// Create a new controller
	c, err := controller.New("blogpost-controller", mgr, controller.Options{Reconciler: r})
	if err != nil {
		return err
	}

	// Watch for changes to BlogPost
	err = c.Watch(&source.Kind{Type: &bloggingv1alpha1.BlogPost{}}, &handler.EnqueueRequestForObject{})
	if err != nil {
		return err
	}

	return nil
}

var _ reconcile.Reconciler = &ReconcileBlogPost{}

// ReconcileBlogPost reconciles a BlogPost object
type ReconcileBlogPost struct {
	client.Client
	scheme *runtime.Scheme
}

// Reconcile reads that state of the cluster for a BlogPost object and makes changes based on the state read
// and what is in the BlogPost.Spec
// TODO(user): Modify this Reconcile function to implement your Controller logic.  The scaffolding writes
// a Deployment as an example
// Automatically generate RBAC rules to allow the Controller to read and write Deployments
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=deployments/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=blogging.dexhorthy.com,resources=blogposts,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=blogging.dexhorthy.com,resources=blogposts/status,verbs=get;update;patch
func (r *ReconcileBlogPost) Reconcile(request reconcile.Request) (reconcile.Result, error) {
	ctx := context.Background()
	// Fetch the BlogPost blogPost
	blogPost := &bloggingv1alpha1.BlogPost{}
	err := r.Get(context.TODO(), request.NamespacedName, blogPost)
	if err != nil {
		if errors.IsNotFound(err) {
			// Object not found, return.  Created objects are automatically garbage collected.
			// For additional cleanup logic use finalizers.
			return reconcile.Result{}, nil
		}
		// Error reading the object - requeue the request.
		return reconcile.Result{}, err
	}

	blog := &bloggingv1alpha1.Blog{}
	err = r.Get(ctx, types.NamespacedName{Name: blogPost.Spec.Blog, Namespace: blogPost.Namespace}, blog)
	if err != nil {
		// Error reading the object, or not parent blog not found - either is a failure
		return reconcile.Result{}, err
	}

	blogContentCM := &corev1.ConfigMap{}
	if err = r.Get(ctx, types.NamespacedName{Name: blogPost.Spec.Blog + "-content", Namespace: blogPost.Namespace}, blogContentCM); err != nil {
		return reconcile.Result{}, err
	}

	if blogContentCM.Data == nil {
		blogContentCM.Data = map[string]string{}
	}
	blogContentCM.Data[blogPost.Name+".md"] = blogPost.Spec.Content
	if err := r.ensureConfigMap(blogContentCM); err != nil {
		return reconcile.Result{}, err
	}

	contentJSON, err := json.Marshal(blogContentCM.Data)
	if err != nil {
		return reconcile.Result{}, err
	}
	h := sha1.New()
	if _, err := io.WriteString(h, string(contentJSON)); err != nil {
		return reconcile.Result{}, err
	}

	hash := fmt.Sprintf("%x", h.Sum(nil))

	blogDeployment := &appsv1.Deployment{}
	if err = r.Get(ctx, types.NamespacedName{Name: blogPost.Spec.Blog, Namespace: blogPost.Namespace}, blogDeployment); err != nil {
		return reconcile.Result{}, err
	}

	var set bool
	for index, envVar := range blogDeployment.Spec.Template.Spec.Containers[0].Env {
		if envVar.Name == "_SITE_CONTENT_HASH" {
			blogDeployment.Spec.Template.Spec.Containers[0].Env[index].Value = hash
			set = true
			break
		}
	}

	if !set {
		blogDeployment.Spec.Template.Spec.Containers[0].Env = append(
			blogDeployment.Spec.Template.Spec.Containers[0].Env,
			corev1.EnvVar{
				Name:  "_SITE_CONTENT_HASH",
				Value: hash,
			},
		)
	}

	if err := r.ensureDeployment(blogDeployment); err != nil {
		return reconcile.Result{}, err
	}

	return reconcile.Result{}, nil
}

func (r *ReconcileBlogPost) ensureConfigMap(cm *corev1.ConfigMap) error {
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
	if !reflect.DeepEqual(cm.Data, found.Data) {
		found.Data = cm.Data
		log.Info("Updating ConfigMap", "namespace", cm.Namespace, "name", cm.Name)
		err = r.Update(context.TODO(), found)
		if err != nil {
			return err
		}
	}
	return nil
}

func (r *ReconcileBlogPost) ensureDeployment(deploy *appsv1.Deployment) error {
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
