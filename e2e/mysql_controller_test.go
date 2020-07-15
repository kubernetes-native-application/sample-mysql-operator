package e2e

import (
	"context"
	"fmt"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	v12 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	"sample-mysql-operator/api/v1alpha1"
	"time"
)

var _ = Describe("MySQL Controller E2E tests", func() {
	Context("Basic", func() {
		It("Should ok", func() {
			By("Creating configMap")
			configMap := &corev1.ConfigMap{
				ObjectMeta: v1.ObjectMeta{
					Name:      "mysql",
					Namespace: "default",
				},
				Data: map[string]string{
					"master.cnf": "[mysqld]\nlog-bin",
					"slave.cnf":  "[mysqld]\nsuper-read-only",
				},
			}
			Expect(k8sClient.Create(context.TODO(), configMap)).Should(Succeed())

			By("Creating mysql")
			toCreate := &v1alpha1.MySQL{
				ObjectMeta: v1.ObjectMeta{
					Name:      "mysql-sample",
					Namespace: "default",
				},
				Spec: v1alpha1.MySQLSpec{
					Replicas:  2,
					OwnerName: "woohyung han",
				},
			}
			Expect(k8sClient.Create(context.TODO(), toCreate)).Should(Succeed())
			defer func() {
				toDelete := &v1alpha1.MySQL{
					ObjectMeta: v1.ObjectMeta{
						Name:      "mysql-sample",
						Namespace: "default",
					},
				}
				By("Deleting mysql")
				Expect(k8sClient.Delete(context.TODO(), toDelete)).Should(Succeed())

				By("Waiting for deleting")
				Eventually(func() bool {
					mysql := &v1alpha1.MySQL{}
					if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "mysql-sample"}, mysql); err != nil {
						return errors.IsNotFound(err)
					}
					return false
				}, 180*time.Second, 5*time.Second).Should(BeTrue())

				By("Deleting configMap")
				Expect(k8sClient.Delete(context.TODO(), configMap)).Should(Succeed())
			}()

			By("Waiting for mysql to be running condition")
			Eventually(func() bool {
				mysql := &v1alpha1.MySQL{}
				if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "mysql-sample"}, mysql); err != nil {
					return false
				}
				fmt.Fprintf(GinkgoWriter, "ConditionTypeRunning: %t\n", mysql.Status.Conditions.IsTrueFor(v1alpha1.ConditionTypeRunning))
				return mysql.Status.Conditions.IsTrueFor(v1alpha1.ConditionTypeRunning)
			}, 300*time.Second, 5*time.Second).Should(BeTrue())

			By("Updating mysql replicas 2->3")
			toUpdate := &v1alpha1.MySQL{}
			Expect(k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "mysql-sample"}, toUpdate)).Should(Succeed())
			toUpdate.Spec.Replicas = 3
			Expect(k8sClient.Update(context.TODO(), toUpdate)).Should(Succeed())

			By("Waiting for mysql replicas count is 3")
			Eventually(func() bool {
				statefulSet := &v12.StatefulSet{}
				if err := k8sClient.Get(context.TODO(), types.NamespacedName{Namespace: "default", Name: "mysql-sample"}, statefulSet); err != nil {
					return false
				}
				fmt.Fprintf(GinkgoWriter, "Replicas: %d\n", statefulSet.Status.Replicas)
				return statefulSet.Status.Replicas == 3
			}, 180*time.Second, 5*time.Second).Should(BeTrue())
		})
	})
})
