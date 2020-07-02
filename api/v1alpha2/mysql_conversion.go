package v1alpha2

import (
	"errors"
	"sample-mysql-operator/api/v1alpha1"
	"sigs.k8s.io/controller-runtime/pkg/conversion"
	"strings"
)

// ConvertTo 는 MySQL(v1alpha2)를 MySQL(v1alpha1)으로 변환한다
func (src *MySQL) ConvertTo(dstRaw conversion.Hub) error {
	dst := dstRaw.(*v1alpha1.MySQL)
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Replicas = src.Spec.Replicas
	dst.Spec.OwnerName = src.Spec.OwnerFirstName + " " + src.Spec.OwnerLastName
	return nil
}

// ConvertFrom 은 MySQL(v1alpha1)을 MySQL(v1alpha2)로 변환한다
func (dst *MySQL) ConvertFrom(srcRaw conversion.Hub) error {
	src := srcRaw.(*v1alpha1.MySQL)
	dst.ObjectMeta = src.ObjectMeta
	dst.Spec.Replicas = src.Spec.Replicas
	name := strings.Split(src.Spec.OwnerName, " ")
	if len(name) != 2 {
		return errors.New("invalid name")
	}
	dst.Spec.OwnerFirstName = name[0]
	dst.Spec.OwnerLastName = name[1]
	return nil
}
