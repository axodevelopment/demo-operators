/*
Copyright 2026.

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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/util/intstr"

	workshopv1 "github.com/axodevelopment/demo-operators/api/v1"
	routev1 "github.com/openshift/api/route/v1"
)

// PaychexReconciler reconciles a Paychex object
type PaychexReconciler struct {
	client.Client
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=workshop.io,resources=paychexes,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=workshop.io,resources=paychexes/status,verbs=get;update;patch
// +kubebuilder:rbac:groups=workshop.io,resources=paychexes/finalizers,verbs=update
// +kubebuilder:rbac:groups=apps,resources=deployments,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=core,resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=route.openshift.io,resources=routes,verbs=get;list;watch;create;update;patch;delete

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Paychex object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.21.0/pkg/reconcile
func (r *PaychexReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	var px workshopv1.Paychex
	if err := r.Get(ctx, req.NamespacedName, &px); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	ns := px.Namespace
	extra := px.Spec.Labels

	desired := []client.Object{
		calleeDeployment(px.Name, ns, extra),
		calleeService(px.Name, ns, extra),
		callerDeployment(px.Name, ns, extra),
		callerService(px.Name, ns, extra),
		callerRoute(px.Name, ns, extra),
	}

	for _, obj := range desired {

		//TODO: note owner
		if err := controllerutil.SetControllerReference(&px, obj, r.Scheme); err != nil {
			return ctrl.Result{}, err
		}

		if err := r.upsert(ctx, obj); err != nil {
			logger.Error(err, "upsert failed", "name", obj.GetName())

			return ctrl.Result{}, err
		}
	}
	return ctrl.Result{}, nil
}

// callee
func calleeDeployment(name, ns string, extra map[string]string) *appsv1.Deployment {
	resourceName := "callee-" + name

	selector := map[string]string{
		"app":     "callee",
		"version": "v1",
		"paychex": name,
	}

	labels := mergeLabels(selector, extra)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: ns, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "callee",
						Image: "quay.io/axodevelopment/grpc_callee:latest",
						Ports: []corev1.ContainerPort{{Name: "grpc", ContainerPort: 50051}},
						Env: []corev1.EnvVar{{
							Name: "POD_NAME",
							ValueFrom: &corev1.EnvVarSource{
								FieldRef: &corev1.ObjectFieldSelector{FieldPath: "metadata.name"},
							},
						}},
					}},
				},
			},
		},
	}
}

func calleeService(name, ns string, extra map[string]string) *corev1.Service {
	resourceName := "callee-" + name

	selector := map[string]string{
		"app":     "callee",
		"paychex": name,
	}

	labels := mergeLabels(selector, extra)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: ns, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{{
				Name: "grpc", Port: 50051, TargetPort: intstr.FromString("grpc"),
			}},
		},
	}
}

// caller
func callerDeployment(name, ns string, extra map[string]string) *appsv1.Deployment {

	resourceName := "caller-" + name

	selector := map[string]string{
		"app":     "caller",
		"version": "v1",
		"paychex": name,
	}

	labels := mergeLabels(selector, extra)

	return &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: ns, Labels: labels},
		Spec: appsv1.DeploymentSpec{
			Replicas: int32Ptr(1),
			Selector: &metav1.LabelSelector{MatchLabels: selector},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{Labels: labels},
				Spec: corev1.PodSpec{
					Containers: []corev1.Container{{
						Name:  "caller",
						Image: "quay.io/axodevelopment/grpc_caller:latest",
						Ports: []corev1.ContainerPort{{Name: "http", ContainerPort: 8080}},
						Env: []corev1.EnvVar{{
							Name:  "CALLEE_ADDR",
							Value: "callee-" + name + ":50051",
						}},
						ReadinessProbe: &corev1.Probe{
							ProbeHandler: corev1.ProbeHandler{
								HTTPGet: &corev1.HTTPGetAction{
									Path: "/healthz", Port: intstr.FromString("http"),
								},
							},
						},
					}},
				},
			},
		},
	}
}

func callerService(name, ns string, extra map[string]string) *corev1.Service {
	resourceName := "caller-" + name

	selector := map[string]string{
		"app":     "caller",
		"paychex": name,
	}

	labels := mergeLabels(selector, extra)

	return &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: ns, Labels: labels},
		Spec: corev1.ServiceSpec{
			Selector: selector,
			Ports: []corev1.ServicePort{{
				Name: "http", Port: 8080, TargetPort: intstr.FromString("http"),
			}},
		},
	}
}

func callerRoute(name, ns string, extra map[string]string) *routev1.Route {
	resourceName := "caller-" + name

	labels := mergeLabels(map[string]string{"app": "caller", "paychex": name}, extra)

	return &routev1.Route{
		ObjectMeta: metav1.ObjectMeta{Name: resourceName, Namespace: ns, Labels: labels},
		Spec: routev1.RouteSpec{
			To:   routev1.RouteTargetReference{Kind: "Service", Name: resourceName},
			Port: &routev1.RoutePort{TargetPort: intstr.FromString("http")},
			TLS: &routev1.TLSConfig{
				Termination:                   routev1.TLSTerminationEdge,
				InsecureEdgeTerminationPolicy: routev1.InsecureEdgeTerminationPolicyAllow,
			},
			WildcardPolicy: routev1.WildcardPolicyNone,
		},
	}
}

func (r *PaychexReconciler) upsert(ctx context.Context, obj client.Object) error {
	existing := obj.DeepCopyObject().(client.Object)
	err := r.Get(ctx, client.ObjectKeyFromObject(obj), existing)

	if apierrors.IsNotFound(err) {
		return r.Create(ctx, obj)
	} else if err != nil {
		return err
	}

	obj.SetResourceVersion(existing.GetResourceVersion())
	return r.Update(ctx, obj)
}

func mergeLabels(base, extra map[string]string) map[string]string {
	out := map[string]string{}
	for k, v := range base {
		out[k] = v
	}
	for k, v := range extra {
		out[k] = v
	}
	return out
}

func int32Ptr(i int32) *int32 { return &i }

// SetupWithManager sets up the controller with the Manager.
func (r *PaychexReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&workshopv1.Paychex{}).
		Owns(&appsv1.Deployment{}).
		Owns(&corev1.Service{}).
		Owns(&routev1.Route{}).
		Complete(r)
}
