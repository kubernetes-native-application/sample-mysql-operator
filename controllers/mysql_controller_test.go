package controllers

import (
	"context"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/types"
	"k8s.io/client-go/kubernetes/scheme"
	"sample-mysql-operator/api/v1alpha1"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client/fake"
	"sigs.k8s.io/controller-runtime/pkg/reconcile"
)

var _ = Describe("mysql reconcile", func() {
	Context("with not created yet", func() {
		var (
			mysql *v1alpha1.MySQL
			r     *MySQLReconciler
		)

		BeforeEach(func() {
			s := scheme.Scheme
			Expect(v1alpha1.AddToScheme(s)).Should(Succeed())
			mysql = &v1alpha1.MySQL{
				ObjectMeta: v1.ObjectMeta{
					Name:      "sample",
					Namespace: "default",
				},
				Spec: v1alpha1.MySQLSpec{
					Replicas:  2,
					OwnerName: "woohyung han",
				},
			}
			objs := []runtime.Object{mysql}
			client := fake.NewFakeClientWithScheme(s, objs...)
			r = &MySQLReconciler{
				Client: client,
				Log:    ctrl.Log.WithName("controllers").WithName("MySQL"),
				Scheme: s,
			}
		})

		It("should succeed", func() {
			_, err := r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "sample",
			}})
			Expect(err).ShouldNot(HaveOccurred())
		})

		It("should create a read service", func() {
			_, err := r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "sample",
			}})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(r.Client.Get(context.TODO(), types.NamespacedName{Name: "sample-read", Namespace: "default"}, &corev1.Service{})).Should(Succeed())
		})

		It("should create a headless service", func() {
			_, err := r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "sample",
			}})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(r.Client.Get(context.TODO(), types.NamespacedName{Name: "sample", Namespace: "default"}, &corev1.Service{})).Should(Succeed())
		})

		It("should create a statefulSet", func() {
			_, err := r.Reconcile(reconcile.Request{NamespacedName: types.NamespacedName{
				Namespace: "default",
				Name:      "sample",
			}})
			Expect(err).ShouldNot(HaveOccurred())
			Expect(r.Client.Get(context.TODO(), types.NamespacedName{Name: "sample", Namespace: "default"}, &v12.StatefulSet{})).Should(Succeed())
		})
	})
})
