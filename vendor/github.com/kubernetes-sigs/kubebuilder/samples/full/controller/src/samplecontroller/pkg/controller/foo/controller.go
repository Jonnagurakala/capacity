/*
Copyright 2017 The Kubernetes Authors.

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

package foo

import (
    "fmt"

    "github.com/golang/glog"
    "github.com/kubernetes-sigs/kubebuilder/pkg/controller"
    "github.com/kubernetes-sigs/kubebuilder/pkg/controller/eventhandlers"
    "github.com/kubernetes-sigs/kubebuilder/pkg/controller/predicates"
    "github.com/kubernetes-sigs/kubebuilder/pkg/controller/types"
    appsv1 "k8s.io/api/apps/v1"
    corev1 "k8s.io/api/core/v1"
    "k8s.io/apimachinery/pkg/api/errors"
    metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
    "k8s.io/apimachinery/pkg/runtime/schema"
    "k8s.io/apimachinery/pkg/util/runtime"
    "k8s.io/client-go/kubernetes/scheme"
    "k8s.io/client-go/tools/record"

    samplecontrollerv1alpha1 "github.com/kubernetes-sigs/kubebuilder/samples/full/controller/src/samplecontroller/pkg/apis/samplecontroller/v1alpha1"
    samplescheme "github.com/kubernetes-sigs/kubebuilder/samples/full/controller/src/samplecontroller/pkg/client/clientset/versioned/scheme"
    "github.com/kubernetes-sigs/kubebuilder/samples/full/controller/src/samplecontroller/pkg/inject/args"
)

const controllerAgentName = "sample-controller"

const (
    // SuccessSynced is used as part of the Event 'reason' when a Foo is synced
    SuccessSynced = "Synced"
    // ErrResourceExists is used as part of the Event 'reason' when a Foo fails
    // to sync due to a Deployment of the same name already existing.
    ErrResourceExists = "ErrResourceExists"

    // MessageResourceExists is the message used for Events when a resource
    // fails to sync due to a Deployment already existing
    MessageResourceExists = "Resource %q already exists and is not managed by Foo"
    // MessageResourceSynced is the message used for an Event fired when a Foo
    // is synced successfully
    MessageResourceSynced = "Foo synced successfully"
)

// Reconcile compares the actual state with the desired, and attempts to
// converge the two. It then updates the Status block of the Foo resource
// with the current status of the resource.
func (bc *FooController) Reconcile(k types.ReconcileKey) error {
    namespace, name := k.Namespace, k.Name
    foo, err := bc.Informers.Samplecontroller().V1alpha1().Foos().Lister().Foos(namespace).Get(name)
    if err != nil {
        // The Foo resource may no longer exist, in which case we stop
        // processing.
        if errors.IsNotFound(err) {
            runtime.HandleError(fmt.Errorf("foo '%s' in work queue no longer exists", k))
            return nil
        }

        return err
    }

    deploymentName := foo.Spec.DeploymentName
    if deploymentName == "" {
        // We choose to absorb the error here as the worker would requeue the
        // resource otherwise. Instead, the next time the resource is updated
        // the resource will be queued again.
        runtime.HandleError(fmt.Errorf("%s: deployment name must be specified", k))
        return nil
    }

    // Get the deployment with the name specified in Foo.spec
    deployment, err := bc.KubernetesInformers.Apps().V1().Deployments().Lister().Deployments(foo.Namespace).Get(deploymentName)
    // If the resource doesn't exist, we'll create it
    if errors.IsNotFound(err) {
        deployment, err = bc.KubernetesClientSet.AppsV1().Deployments(foo.Namespace).Create(newDeployment(foo))
    }

    // If an error occurs during Get/Create, we'll requeue the item so we can
    // attempt processing again later. This could have been caused by a
    // temporary network failure, or any other transient reason.
    if err != nil {
        return err
    }

    // If the Deployment is not controlled by this Foo resource, we should log
    // a warning to the event recorder and ret
    if !metav1.IsControlledBy(deployment, foo) {
        msg := fmt.Sprintf(MessageResourceExists, deployment.Name)
        bc.recorder.Event(foo, corev1.EventTypeWarning, ErrResourceExists, msg)
        return fmt.Errorf(msg)
    }

    // If this number of the replicas on the Foo resource is specified, and the
    // number does not equal the current desired replicas on the Deployment, we
    // should update the Deployment resource.
    if foo.Spec.Replicas != nil && *foo.Spec.Replicas != *deployment.Spec.Replicas {
        glog.V(4).Infof("Foo %s replicas: %d, deployment replicas: %d", name, *foo.Spec.Replicas, *deployment.Spec.Replicas)
        deployment, err = bc.KubernetesClientSet.AppsV1().Deployments(foo.Namespace).Update(newDeployment(foo))
    }

    // If an error occurs during Update, we'll requeue the item so we can
    // attempt processing again later. THis could have been caused by a
    // temporary network failure, or any other transient reason.
    if err != nil {
        return err
    }

    // Finally, we update the status block of the Foo resource to reflect the
    // current state of the world
    err = bc.updateFooStatus(foo, deployment)
    if err != nil {
        return err
    }

    bc.recorder.Event(foo, corev1.EventTypeNormal, SuccessSynced, MessageResourceSynced)
    return nil
}

// FooController is the controller implementation for Foo resources
// +controller:group=samplecontroller,version=v1alpha1,kind=Foo,resource=foos
// +informers:group=apps,version=v1,kind=Deployment
// +rbac:rbac:groups=apps,resources=Deployment,verbs=get;list;watch;create;update;patch;delete
type FooController struct {
    args.InjectArgs

    // recorder is an event recorder for recording Event resources to the
    // Kubernetes API.
    recorder record.EventRecorder
}

// ProvideController provides a controller that will be run at startup.  Kubebuilder will use codegeneration
// to automatically register this controller in the inject package
func ProvideController(iargs args.InjectArgs) (*controller.GenericController, error) {
    samplescheme.AddToScheme(scheme.Scheme)

    bc := &FooController{
        InjectArgs: iargs,
        recorder: iargs.CreateRecorder(controllerAgentName),
    }

    // Create a new controller that will call FooController.Reconcile on changes to Foos
    gc := &controller.GenericController{
        Name:             controllerAgentName,
        Reconcile:        bc.Reconcile,
        InformerRegistry: iargs.ControllerManager,
    }

    glog.Info("Setting up event handlers")
    if err := gc.Watch(&samplecontrollerv1alpha1.Foo{}); err != nil {
        return gc, err
    }

    // Set up an event handler for when Deployment resources change. This
    // handler will lookup the owner of the given Deployment, and if it is
    // owned by a Foo resource will enqueue that Foo resource for
    // processing. This way, we don't need to implement custom logic for
    // handling Deployment resources. More info on this pattern:
    // https://github.com/kubernetes/community/blob/8cafef897a22026d42f5e5bb3f104febe7e29830/contributors/devel/controllers.md
    if err := gc.WatchControllerOf(&appsv1.Deployment{}, eventhandlers.Path{bc.LookupFoo},
        predicates.ResourceVersionChanged); err != nil {
        return gc, err
    }

    return gc, nil
}

// LookupFoo looksup a Foo from the lister
func (bc FooController) LookupFoo(r types.ReconcileKey) (interface{}, error) {
    return bc.Informers.Samplecontroller().V1alpha1().Foos().Lister().Foos(r.Namespace).Get(r.Name)
}

func (bc *FooController) updateFooStatus(foo *samplecontrollerv1alpha1.Foo, deployment *appsv1.Deployment) error {
    // NEVER modify objects from the store. It's a read-only, local cache.
    // You can use DeepCopy() to make a deep copy of original object and modify this copy
    // Or create a copy manually for better performance
    fooCopy := foo.DeepCopy()
    fooCopy.Status.AvailableReplicas = deployment.Status.AvailableReplicas
    // Until #38113 is merged, we must use Update instead of UpdateStatus to
    // update the Status block of the Foo resource. UpdateStatus will not
    // allow changes to the Spec of the resource, which is ideal for ensuring
    // nothing other than resource status has been updated.
    _, err := bc.Clientset.SamplecontrollerV1alpha1().Foos(foo.Namespace).Update(fooCopy)
    return err
}

// newDeployment creates a new Deployment for a Foo resource. It also sets
// the appropriate OwnerReferences on the resource so handleObject can discover
// the Foo resource that 'owns' it.
func newDeployment(foo *samplecontrollerv1alpha1.Foo) *appsv1.Deployment {
    labels := map[string]string{
        "app":        "nginx",
        "controller": foo.Name,
    }
    return &appsv1.Deployment{
        ObjectMeta: metav1.ObjectMeta{
            Name:      foo.Spec.DeploymentName,
            Namespace: foo.Namespace,
            OwnerReferences: []metav1.OwnerReference{
                *metav1.NewControllerRef(foo, schema.GroupVersionKind{
                    Group:   samplecontrollerv1alpha1.SchemeGroupVersion.Group,
                    Version: samplecontrollerv1alpha1.SchemeGroupVersion.Version,
                    Kind:    "Foo",
                }),
            },
        },
        Spec: appsv1.DeploymentSpec{
            Replicas: foo.Spec.Replicas,
            Selector: &metav1.LabelSelector{
                MatchLabels: labels,
            },
            Template: corev1.PodTemplateSpec{
                ObjectMeta: metav1.ObjectMeta{
                    Labels: labels,
                },
                Spec: corev1.PodSpec{
                    Containers: []corev1.Container{
                        {
                            Name:  "nginx",
                            Image: "nginx:latest",
                        },
                    },
                },
            },
        },
    }
}
