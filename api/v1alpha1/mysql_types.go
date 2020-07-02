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
	"github.com/woohhan/kubebuilder-util/pkg/condition"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// MySQLSpec defines the desired state of MySQL
type MySQLSpec struct {
	// Replicas 는 MySQL의 복제 개수를 나타낸다
	// +kubebuilder:validation:Maximum=5
	// +kubebuilder:validation:Minimum=1
	Replicas int32 `json:"replicas"`
	// OwnerName 은 이 MySQL의 주인을 나타낸다. 반드시 [first name] [last name] 형태로 입력되어야 한다.
	OwnerName string `json:"ownerName,omitempty"`
}

// MySQLStatus defines the observed state of MySQL
type MySQLStatus struct {
	// Conditions represent the latest available observations of an object's state
	Conditions condition.Conditions `json:"conditions"`
}

const (
	// ConditionTypeRunning 은 MySQL이 동작하고 있는지를 나타낸다
	ConditionTypeRunning condition.ConditionType = "Running"
)

// +kubebuilder:object:root=true
// +kubebuilder:subresource:status

// MySQL is the Schema for the mysqls API
type MySQL struct {
	metav1.TypeMeta   `json:",inline"`
	metav1.ObjectMeta `json:"metadata,omitempty"`

	Spec   MySQLSpec   `json:"spec,omitempty"`
	Status MySQLStatus `json:"status,omitempty"`
}

// +kubebuilder:object:root=true

// MySQLList contains a list of MySQL
type MySQLList struct {
	metav1.TypeMeta `json:",inline"`
	metav1.ListMeta `json:"metadata,omitempty"`
	Items           []MySQL `json:"items"`
}

func init() {
	SchemeBuilder.Register(&MySQL{}, &MySQLList{})
}
