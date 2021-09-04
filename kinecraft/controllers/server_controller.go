/*
Copyright 2021.

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

package controllers

import (
	"context"
	"fmt"
	"strings"

	"github.com/go-logr/logr"
	minecraftv1 "github.com/kirankothule/kinecraft/api/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
)

// ServerReconciler reconciles a Server object
type ServerReconciler struct {
	client.Client
	Scheme *runtime.Scheme
	Log    logr.Logger
}

//+kubebuilder:rbac:groups=minecraft.kk.io,resources=servers,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=minecraft.kk.io,resources=servers/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=minecraft.kk.io,resources=servers/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the Server object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.8.3/pkg/reconcile

// TODO(KK): list RBAC stuff
func (r *ServerReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	_ = log.FromContext(ctx)
	log := r.Log.WithValues("server", req.NamespacedName)
	// your logic here

	var mcServer minecraftv1.Server
	if err := r.Get(ctx, req.NamespacedName, &mcServer); err != nil {
		log.Error(err, "unable to fetch Server")
		// we'll ignore not-found errors, since they can't be fixed by an immediate
		// requeue (we'll need to wait for a new notification), and we can get them
		// on deleted requests.
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	// TODO(KK): List out the pods that belongs to this server and update the status field

	pod, err := r.constructPod(&mcServer)
	if err != nil {
		return ctrl.Result{}, err
	}
	// ...and create it on the cluster
	if err := r.Create(ctx, pod); err != nil {
		log.Error(err, "unable to create pod for Server", "pod", pod)
		return ctrl.Result{}, err
	}

	log.V(1).Info("created pod for Server run", "pod", pod)

	return ctrl.Result{}, nil
}

func (r *ServerReconciler) constructPod(s *minecraftv1.Server) (*corev1.Pod, error) {
	name := fmt.Sprintf("mc-%s", s.Name)
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:        name,
			Namespace:   s.Namespace,
			Labels:      make(map[string]string),
			Annotations: make(map[string]string),
		},
		Spec: corev1.PodSpec{
			Containers: []corev1.Container{
				corev1.Container{
					Name:  "mincraft-server",
					Image: "itzg/minecraft-server",
					Ports: []corev1.ContainerPort{
						corev1.ContainerPort{
							Name:          "minecraft",
							ContainerPort: 25565,
							Protocol:      corev1.ProtocolTCP,
						},
					},
					Env: []corev1.EnvVar{},
				},
			},
		},
		Status: corev1.PodStatus{},
	}

	addEnv := func(key, value string) {
		pod.Spec.Containers[0].Env = append(pod.Spec.Containers[0].Env, corev1.EnvVar{Name: key, Value: value})
	}
	bool2String := func(b bool) string {
		if b {
			return "TRUE"
		} else {
			return "FALSE"
		}
	}
	addEnv("EULA", bool2String(s.Spec.EULA))
	addEnv("TYPE", s.Spec.ServerType)
	addEnv("SERVER_NAME", s.Spec.ServerName)
	addEnv("OPS", strings.Join(s.Spec.Ops, ","))
	addEnv("WHITELIST", strings.Join(s.Spec.AllowList, ","))

	if err := ctrl.SetControllerReference(pod, s, r.Scheme); err != nil {
		return pod, err
	}
	return pod, nil

}

// SetupWithManager sets up the controller with the Manager.
func (r *ServerReconciler) SetupWithManager(mgr ctrl.Manager) error {
	// TODO(kk): Make sure that we are getting kicked on pod changes also
	return ctrl.NewControllerManagedBy(mgr).
		For(&minecraftv1.Server{}).
		Complete(r)
}
