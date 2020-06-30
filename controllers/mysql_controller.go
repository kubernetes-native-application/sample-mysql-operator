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
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
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
	if err := r.syncHeadlessService(mysql); err != nil {
		return ctrl.Result{}, err
	}
	if err := r.syncStatefulSet(mysql); err != nil {
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
		r.Log.Info("Could not find read service. Create a new one")
		return createReadService(r, mysql)
	}
	return nil
}

func createReadService(r *MySQLReconciler, mysql *mysqlv1alpha1.MySQL) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
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

func (r *MySQLReconciler) syncHeadlessService(mysql *mysqlv1alpha1.MySQL) error {
	svc := &corev1.Service{}
	if err := r.Get(context.TODO(), types.NamespacedName{Namespace: mysql.Namespace, Name: mysql.Name}, svc); err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		r.Log.Info("Could not find headless service. Create a new one")
		return createHeadlessService(r, mysql)
	}
	return nil
}

func createHeadlessService(r *MySQLReconciler, mysql *mysqlv1alpha1.MySQL) error {
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Namespace: mysql.Namespace,
			Name:      mysql.Name,
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
			ClusterIP: "None",
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

func (r *MySQLReconciler) syncStatefulSet(mysql *mysqlv1alpha1.MySQL) error {
	// 클라이언트로 MySQL 스테이트풀셋 객체를 가져온다
	sf := &appsv1.StatefulSet{}
	err := r.Get(context.TODO(), types.NamespacedName{Namespace: mysql.Namespace, Name: mysql.Name}, sf)
	if err != nil {
		if !errors.IsNotFound(err) {
			return err
		}
		// MySQL 스테이트풀셋을 생성한다
		r.Log.Info("Could not find mysql StatefulSet. Create a new one")
		return createStatefulSet(r, mysql)
	}
	// 스테이트풀셋의 레플리카 수가 변경된 경우 업데이트한다
	if mysql.Spec.Replicas != *sf.Spec.Replicas {
		r.Log.Info("Update replica size")
		clonedStatefulSet := sf.DeepCopy()
		*clonedStatefulSet.Spec.Replicas = mysql.Spec.Replicas
		if err := r.Update(context.TODO(), clonedStatefulSet); err != nil {
			return err
		}
		return nil
	}
	return nil
}

func createStatefulSet(r *MySQLReconciler, mysql *mysqlv1alpha1.MySQL) error {
	volumeMount := []corev1.VolumeMount{
		{
			Name:      "data",
			MountPath: "/var/lib/mysql",
			SubPath:   "mysql",
		},
		{
			Name:      "conf",
			MountPath: "/etc/mysql/conf.d",
		},
	}
	statefulSet := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      mysql.Name,
			Namespace: mysql.Namespace,
		},
		Spec: appsv1.StatefulSetSpec{
			Replicas: &mysql.Spec.Replicas,
			Selector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"app": mysql.Name,
				},
			},
			Template: corev1.PodTemplateSpec{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{
						"app": mysql.Name,
					},
				},
				Spec: corev1.PodSpec{
					Volumes: []corev1.Volume{
						{
							Name: "conf",
							VolumeSource: corev1.VolumeSource{
								EmptyDir: &corev1.EmptyDirVolumeSource{},
							},
						},
						{
							Name: "config-map",
							VolumeSource: corev1.VolumeSource{
								ConfigMap: &corev1.ConfigMapVolumeSource{
									LocalObjectReference: corev1.LocalObjectReference{
										Name: "mysql",
									},
								},
							},
						},
					},
					InitContainers: []corev1.Container{
						{
							Name:  "init-mysql",
							Image: "quay.io/sample-mysql-operator/init-mysql:latest",
							VolumeMounts: []corev1.VolumeMount{
								{
									Name:      "conf",
									MountPath: "/mnt/conf.d",
								},
								{
									Name:      "config-map",
									MountPath: "/mnt/config-map",
								},
							},
						},
						{
							Name:         "clone-mysql",
							Image:        "quay.io/sample-mysql-operator/clone-mysql:latest",
							VolumeMounts: volumeMount,
							Env: []corev1.EnvVar{
								{
									Name:  "POD_NAME",
									Value: mysql.Name,
								},
								{
									Name:  "SVC_NAME",
									Value: mysql.Name,
								},
							},
						},
					},
					Containers: []corev1.Container{
						{
							Name:  "mysql",
							Image: "mysql:5.7",
							Env: []corev1.EnvVar{
								{
									Name:  "MYSQL_ALLOW_EMPTY_PASSWORD",
									Value: "1",
								},
							},
							VolumeMounts: volumeMount,
							LivenessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mysqladmin", "ping",
										},
									},
								},
								InitialDelaySeconds: 30,
								TimeoutSeconds:      5,
								PeriodSeconds:       10,
							},
							ReadinessProbe: &corev1.Probe{
								Handler: corev1.Handler{
									Exec: &corev1.ExecAction{
										Command: []string{
											"mysql", "-h", "127.0.0.1", "-e", "SELECT 1",
										},
									},
								},
								InitialDelaySeconds: 5,
								TimeoutSeconds:      1,
								PeriodSeconds:       2,
							},
							Ports: []corev1.ContainerPort{
								{
									Name:          "mysql",
									ContainerPort: 3306,
								},
							},
						},
						{
							Name:  "xtrabackup",
							Image: "quay.io/sample-mysql-operator/xtrabackup:latest",
							Ports: []corev1.ContainerPort{
								{
									Name:          "xtrabackup",
									ContainerPort: 3307,
								},
							},
							VolumeMounts: volumeMount,
							Env: []corev1.EnvVar{
								{
									Name:  "POD_NAME",
									Value: mysql.Name,
								},
								{
									Name:  "SVC_NAME",
									Value: mysql.Name,
								},
							},
						},
					},
				},
			},
			VolumeClaimTemplates: []corev1.PersistentVolumeClaim{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name: "data",
					},
					Spec: corev1.PersistentVolumeClaimSpec{
						AccessModes: []corev1.PersistentVolumeAccessMode{
							"ReadWriteOnce",
						},
						Resources: corev1.ResourceRequirements{
							Requests: map[corev1.ResourceName]resource.Quantity{
								corev1.ResourceStorage: resource.MustParse("2Gi")},
						},
					},
				},
			},
			ServiceName: mysql.Name,
		},
	}
	if err := controllerutil.SetControllerReference(mysql, statefulSet, r.Scheme); err != nil {
		return err
	}
	if err := r.Create(context.TODO(), statefulSet); err != nil && !errors.IsAlreadyExists(err) {
		return err
	}
	return nil
}

func (r *MySQLReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&mysqlv1alpha1.MySQL{}).
		Owns(&corev1.Service{}).
		Owns(&appsv1.StatefulSet{}).
		Complete(r)
}
