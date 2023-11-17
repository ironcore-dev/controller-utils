// Copyright 2022 IronCore authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package conditionutils_test

import (
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	. "github.com/ironcore-dev/controller-utils/conditionutils"
)

var _ = Describe("Alias", func() {
	var (
		now   time.Time
		cond  appsv1.DeploymentCondition
		conds []appsv1.DeploymentCondition
	)
	BeforeEach(func() {
		now = time.Now()
		cond = appsv1.DeploymentCondition{
			Type:               appsv1.DeploymentAvailable,
			Status:             corev1.ConditionTrue,
			LastUpdateTime:     metav1.Unix(2, 0),
			LastTransitionTime: metav1.Unix(1, 0),
			Reason:             "MinimumReplicasAvailable",
			Message:            "ReplicaSet \"foo\" has successfully progressed.",
		}
		conds = []appsv1.DeploymentCondition{cond}
	})

	Describe("Update", func() {
		It("should update a condition", func() {
			Expect(Update(&cond, UpdateStatus(corev1.ConditionFalse))).To(Succeed())
			Expect(cond.Status).To(Equal(corev1.ConditionFalse))
			Expect(cond.LastTransitionTime.Time).To(BeTemporally(">=", now))
			Expect(cond.LastUpdateTime.Time).To(BeTemporally(">=", now))
		})

		It("should error if it cannot update a condition", func() {
			Expect(Update(1)).To(HaveOccurred())
		})
	})

	Describe("MustUpdate", func() {
		It("should update a condition", func() {
			Expect(func() { MustUpdate(&cond, UpdateStatus(corev1.ConditionFalse)) }).NotTo(Panic())
			Expect(cond.Status).To(Equal(corev1.ConditionFalse))
			Expect(cond.LastTransitionTime.Time).To(BeTemporally(">=", now))
			Expect(cond.LastUpdateTime.Time).To(BeTemporally(">=", now))
		})

		It("should panic if it cannot update a condition", func() {
			Expect(func() { MustUpdate(1) }).To(Panic())
		})
	})

	Describe("UpdateSlice", func() {
		It("should update the condition slice", func() {
			Expect(UpdateSlice(&conds, string(appsv1.DeploymentAvailable),
				UpdateStatus(corev1.ConditionFalse),
			)).NotTo(HaveOccurred())
			Expect(conds[0].Status).To(Equal(corev1.ConditionFalse))
			Expect(conds[0].LastTransitionTime.Time).To(BeTemporally(">=", now))
			Expect(conds[0].LastUpdateTime.Time).To(BeTemporally(">=", now))
		})

		It("should error if it cannot update the condition slice", func() {
			Expect(UpdateSlice(1, "foo")).To(HaveOccurred())
		})
	})

	Describe("MustUpdateSlice", func() {
		It("should update the condition slice", func() {
			Expect(func() {
				MustUpdateSlice(&conds, string(appsv1.DeploymentAvailable),
					UpdateStatus(corev1.ConditionFalse),
				)
			}).NotTo(Panic())
			Expect(conds[0].Status).To(Equal(corev1.ConditionFalse))
			Expect(conds[0].LastTransitionTime.Time).To(BeTemporally(">=", now))
			Expect(conds[0].LastUpdateTime.Time).To(BeTemporally(">=", now))
		})

		It("should panic if it cannot update the condition slice", func() {
			Expect(func() { MustUpdateSlice(1, "foo") }).To(Panic())
		})
	})

	Describe("FindSliceIndex", func() {
		It("should find the slice index", func() {
			idx, err := FindSliceIndex(conds, string(appsv1.DeploymentAvailable))
			Expect(err).NotTo(HaveOccurred())
			Expect(idx).To(Equal(0))
		})

		It("should error if it cannot find the slice index", func() {
			_, err := FindSliceIndex(1, "foo")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MustFindSliceIndex", func() {
		It("should find the slice index", func() {
			idx := MustFindSliceIndex(conds, string(appsv1.DeploymentAvailable))
			Expect(idx).To(Equal(0))
		})

		It("should panic if it cannot find the slice index", func() {
			Expect(func() { MustFindSliceIndex(1, "foo") }).To(Panic())
		})
	})

	Describe("FindSlice", func() {
		It("should find the slice", func() {
			var actual appsv1.DeploymentCondition
			ok, err := FindSlice(conds, string(appsv1.DeploymentAvailable), &actual)
			Expect(err).NotTo(HaveOccurred())
			Expect(ok).To(BeTrue())
			Expect(actual).To(Equal(cond))
		})

		It("should error if it cannot find the slice", func() {
			_, err := FindSlice(1, "foo", 1)
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MustFindSlice", func() {
		It("should find the slice", func() {
			var actual appsv1.DeploymentCondition
			ok := MustFindSlice(conds, string(appsv1.DeploymentAvailable), &actual)
			Expect(ok).To(BeTrue())
			Expect(actual).To(Equal(cond))
		})

		It("should panic if it cannot find the slice", func() {
			Expect(func() { MustFindSlice(1, "foo", 1) }).To(Panic())
		})
	})

	Describe("FindSliceStatus", func() {
		It("should find the slice status", func() {
			status, err := FindSliceStatus(conds, string(appsv1.DeploymentAvailable))
			Expect(err).NotTo(HaveOccurred())
			Expect(status).To(Equal(corev1.ConditionTrue))
		})

		It("should error if it cannot find the slice status", func() {
			_, err := FindSliceStatus(1, "foo")
			Expect(err).To(HaveOccurred())
		})
	})

	Describe("MustFindSliceStatus", func() {
		It("should find the slice status", func() {
			status := MustFindSliceStatus(conds, string(appsv1.DeploymentAvailable))
			Expect(status).To(Equal(corev1.ConditionTrue))
		})

		It("should panic if it cannot find the slice status", func() {
			Expect(func() { MustFindSliceStatus(1, "foo") }).To(Panic())
		})
	})
})
