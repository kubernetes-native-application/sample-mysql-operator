/*


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
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"

	"github.com/go-logr/logr"
	"k8s.io/apimachinery/pkg/runtime"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"

	mysqlv1alpha1 "sample-mysql-operator/api/v1alpha1"
)

// MySQLReconciler reconciles a MySQL object
type MySQLReconciler struct {
	client.Client
	Log    logr.Logger
	Scheme *runtime.Scheme
}

// +kubebuilder:rbac:groups=mysql.sample.com,resources=mysqls,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=mysql.sample.com,resources=mysqls/status,verbs=get;update;patch
// +kubebuilder:rbac:groups="",resources=services,verbs=get;list;watch;create;update;patch;delete
// +kubebuilder:rbac:groups=apps,resources=statefulsets,verbs=get;list;watch;create;update;patch;delete

func (r *MySQLReconciler) Reconcile(req ctrl.Request) (ctrl.Result, error) {
	r.Log = r.Log.WithValues("mysql", req.NamespacedName)

	mysql := &mysqlv1alpha1.MySQL{}
	if err := r.Get(context.TODO(), req.NamespacedName, mysql); err != nil {
		return ctrl.Result{}, client.IgnoreNotFound(err)
	}

	if err := r.syncReadService(mysql); err != nil {
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *MySQLReconciler) syncReadService(mysql *mysqlv1alpha1.MySQL) error {
	svc := &corev1.Service{}
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: mysql.Namespace, Name: mysql.Name + "-read"}, svc); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		// 서비스가 없는 경우, 서비스를 생성한다
		r.Log.Info("Could not find mysql-read service. Create a new one")
		return createReadService(r, mysql)
	}
	return nil
}

func createReadService(r *MySQLReconciler, mysql *mysqlv1alpha1.MySQL) error {
	svc := &corev1.Service{
		ObjectMeta: v1.ObjectMeta{
			Namespace: mysql.Namespace,
			Name:      mysql.Name + "-read",
		},
		Spec: corev1.ServiceSpec{
			Ports: []corev1.ServicePort{
				{
					Name: "mysql",
					Port: 3306,
				},
			},
			Selector: map[string]string{
				"app": mysql.Name,
			},
		},
	}
	if err := controllerutil.SetControllerReference(mysql, svc, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(context.TODO(), svc); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *MySQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQL{}).
		Owns(&corev1.Service{}).
		Complete(r)
}
