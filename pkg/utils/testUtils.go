// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.

package utils

import (
	"github.com/azure/kaito/api/v1alpha1"
	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
)

const (
	LabelKeyNvidia    = "accelerator"
	LabelValueNvidia  = "nvidia"
	CapacityNvidiaGPU = "nvidia.com/gpu"
)

var (
	MockWorkspaceDistributedModel = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apps": "test",
				},
			},
		},
		Inference: v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-distributed-model",
				},
			},
		},
	}
)

var (
	MockWorkspaceWithPreset = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apps": "test",
				},
			},
		},
		Inference: v1alpha1.InferenceSpec{
			Preset: &v1alpha1.PresetSpec{
				PresetMeta: v1alpha1.PresetMeta{
					Name: "test-model",
				},
			},
		},
	}
)

var (
	MockWorkspaceWithInferenceTemplate = &v1alpha1.Workspace{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "testWorkspace",
			Namespace: "kaito",
		},
		Resource: v1alpha1.ResourceSpec{
			Count:        &gpuNodeCount,
			InstanceType: "Standard_NC12s_v3",
			LabelSelector: &metav1.LabelSelector{
				MatchLabels: map[string]string{
					"apps": "test",
				},
			},
		},
		Inference: v1alpha1.InferenceSpec{
			Template: &corev1.PodTemplateSpec{},
		},
	}
)

var (
	MockNodeList = &corev1.NodeList{
		Items: nodes,
	}
)

var (
	nodes = []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node1",
				Labels: map[string]string{
					corev1.LabelInstanceTypeStable: "Standard_NC12s_v3",
					LabelKeyNvidia:                 LabelValueNvidia,
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionTrue,
					},
				},
				Capacity: corev1.ResourceList{
					CapacityNvidiaGPU: resource.MustParse("1"),
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node2",
				Labels: map[string]string{
					corev1.LabelInstanceTypeStable: "Wrong_Instance_Type",
				},
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node3",
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{
						Type:   corev1.NodeReady,
						Status: corev1.ConditionFalse,
					},
				},
			},
		},
	}
)

var (
	gpuNodeCount = 1
)

var (
	machineLabels = map[string]string{
		"karpenter.sh/provisioner-name": "default",
		"kaito.sh/workspace":            "none",
	}
)

var (
	MockNodeClaim = v1beta1.NodeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:   "testmachine",
			Labels: machineLabels,
		},
		Spec: v1beta1.NodeClaimSpec{
			Requirements: []corev1.NodeSelectorRequirement{
				{
					Key:      corev1.LabelInstanceTypeStable,
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"Standard_NC12s_v3"},
				},
				{
					Key:      "karpenter.sh/provisioner-name",
					Operator: corev1.NodeSelectorOpIn,
					Values:   []string{"default"},
				},
			},
		},
	}
)

var (
	MockNodeClaimList = &v1beta1.NodeClaimList{
		Items: []v1beta1.NodeClaim{
			MockNodeClaim,
		},
	}
)

func NewTestScheme() *runtime.Scheme {
	testScheme := runtime.NewScheme()
	_ = appsv1.AddToScheme(testScheme)
	return testScheme
}

func NotFoundError() error {
	return &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonNotFound}}
}

func IsAlreadyExistsError() error {
	return &apierrors.StatusError{ErrStatus: metav1.Status{Reason: metav1.StatusReasonAlreadyExists}}
}
