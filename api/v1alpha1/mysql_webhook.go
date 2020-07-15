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

package v1alpha1

import (
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/util/validation/field"
	ctrl "sigs.k8s.io/controller-runtime"
	logf "sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/controller-runtime/pkg/webhook"
)

// log is for logging in this package.
var mysqllog = logf.Log.WithName("mysql-resource")

// SetupWebhookWithManager setup new webhook
func (r *MySQL) SetupWebhookWithManager(mgr ctrl.Manager) error {
	return ctrl.NewWebhookManagedBy(mgr).
		For(r).
		Complete()
}

// +kubebuilder:webhook:path=/mutate-mysql-sample-com-v1alpha1-mysql,mutating=true,failurePolicy=fail,groups=mysql.sample.com,resources=mysqls,verbs=create;update,versions=v1alpha1,name=mmysql.kb.io

var _ webhook.Defaulter = &MySQL{}

// Default implements webhook.Defaulter so a webhook will be registered for the type
func (r *MySQL) Default() {
	mysqllog.Info("----- Start Default()", "name", r.Name)

	// 오너 이름이 설정되어 있지 않으면 no body로 설정한다
	if r.Spec.OwnerName == "" {
		r.Spec.OwnerName = "no body"
	}
}

// +kubebuilder:webhook:verbs=create;update,path=/validate-mysql-sample-com-v1alpha1-mysql,mutating=false,failurePolicy=fail,groups=mysql.sample.com,resources=mysqls,versions=v1alpha1,name=vmysql.kb.io

var _ webhook.Validator = &MySQL{}

// ValidateCreate implements webhook.Validator so a webhook will be registered for the type
func (r *MySQL) ValidateCreate() error {
	mysqllog.Info("----- Start ValidateCreate()", "name", r.Name)
	return validateMysql(r)
}

// ValidateUpdate implements webhook.Validator so a webhook will be registered for the type
func (r *MySQL) ValidateUpdate(old runtime.Object) error {
	mysqllog.Info("----- Start ValidateUpdate()", "name", r.Name)
	return validateMysql(r)
}

func validateMysql(mysql *MySQL) error {
	var allErrs field.ErrorList
	if err := validateName(mysql.Spec.OwnerName); err != nil {
		allErrs = append(allErrs, err)
	}
	if len(allErrs) != 0 {
		return apierrors.NewInvalid(schema.GroupKind{Group: "mysql.sample.com", Kind: "MySQL"}, mysql.Name, allErrs)
	}
	return nil
}

func validateName(name string) *field.Error {
	spaceNum := 0
	for _, char := range name {
		if char == ' ' {
			spaceNum++
		}
	}
	if spaceNum != 1 {
		return field.Invalid(field.NewPath("spec"), name, "OwnerName must be form of [first name] [last name]")
	}
	return nil
}

// ValidateDelete implements webhook.Validator so a webhook will be registered for the type
func (r *MySQL) ValidateDelete() error {
	mysqllog.Info("validate delete", "name", r.Name)

	// TODO(user): fill in your validation logic upon object deletion.
	return nil
}
