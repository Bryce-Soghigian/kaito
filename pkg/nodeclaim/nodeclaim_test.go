// Copyright (c) Microsoft Corporation.
// Licensed under the MIT license.
package nodeclaim

import (
	"context"
	"errors"
	"testing"

	"sigs.k8s.io/karpenter/pkg/apis/v1beta1"
	"github.com/azure/kaito/pkg/utils"
	"github.com/stretchr/testify/mock"
	"gotest.tools/assert"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/types"
	"knative.dev/pkg/apis"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

func TestCreateNodeClaim(t *testing.T) {
	testcases := map[string]struct {
		callMocks         func(c *utils.MockClient)
		machineConditions apis.Conditions
		expectedError     error
	}{
		"NodeClaim creation fails": {
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(errors.New("Failed to create machine"))
			},
			expectedError: errors.New("Failed to create machine"),
		},
		"NodeClaim creation fails because SKU is not available": {
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			machineConditions: apis.Conditions{
				{
					Type:    v1beta1.Launched,
					Status:  corev1.ConditionFalse,
					Message: ErrorInstanceTypesUnavailable,
				},
			},
			expectedError: errors.New(ErrorInstanceTypesUnavailable),
		},
		"A machine is successfully created": {
			callMocks: func(c *utils.MockClient) {
				c.On("Create", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			machineConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := utils.NewClient()
			tc.callMocks(mockClient)

			mockNodeClaim := &utils.MockNodeClaim
			mockNodeClaim.Status.Conditions = tc.machineConditions

			err := CreateNodeClaim(context.Background(), mockNodeClaim, mockClient)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestWaitForPendingNodeClaims(t *testing.T) {
	testcases := map[string]struct {
		callMocks         func(c *utils.MockClient)
		machineConditions apis.Conditions
		expectedError     error
	}{
		"Fail to list machines because associated machines cannot be retrieved": {
			callMocks: func(c *utils.MockClient) {
				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(errors.New("Failed to retrieve machines"))
			},
			expectedError: errors.New("Failed to retrieve machines"),
		},
		"Fail to list machines because machine status cannot be retrieved": {
			callMocks: func(c *utils.MockClient) {
				machineList := utils.MockNodeClaimList
				relevantMap := c.CreateMapWithType(machineList)
				c.CreateOrUpdateObjectInMap(&utils.MockNodeClaim)

				//insert machine objects into the map
				for _, obj := range utils.MockNodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(errors.New("Fail to get machine"))
			},
			machineConditions: apis.Conditions{
				{
					Type:   v1beta1.Initialized,
					Status: corev1.ConditionFalse,
				},
			},
			expectedError: errors.New("Fail to get machine"),
		},
		"Successfully waits for all pending machines": {
			callMocks: func(c *utils.MockClient) {
				machineList := utils.MockNodeClaimList
				relevantMap := c.CreateMapWithType(machineList)
				c.CreateOrUpdateObjectInMap(&utils.MockNodeClaim)

				//insert machine objects into the map
				for _, obj := range utils.MockNodeClaimList.Items {
					m := obj
					objKey := client.ObjectKeyFromObject(&m)

					relevantMap[objKey] = &m
				}

				c.On("List", mock.IsType(context.Background()), mock.IsType(&v1beta1.NodeClaimList{}), mock.Anything).Return(nil)
				c.On("Get", mock.IsType(context.Background()), mock.Anything, mock.IsType(&v1beta1.NodeClaim{}), mock.Anything).Return(nil)
			},
			machineConditions: apis.Conditions{
				{
					Type:   apis.ConditionReady,
					Status: corev1.ConditionTrue,
				},
			},
			expectedError: nil,
		},
	}

	for k, tc := range testcases {
		t.Run(k, func(t *testing.T) {
			mockClient := utils.NewClient()
			tc.callMocks(mockClient)

			mockNodeClaim := &v1beta1.NodeClaim{}

			mockClient.UpdateCb = func(key types.NamespacedName) {
				mockClient.GetObjectFromMap(mockNodeClaim, key)
				mockNodeClaim.Status.Conditions = tc.machineConditions
				mockClient.CreateOrUpdateObjectInMap(mockNodeClaim)
			}

			err := WaitForPendingNodeClaims(context.Background(), utils.MockWorkspaceWithPreset, mockClient)
			if tc.expectedError == nil {
				assert.Check(t, err == nil, "Not expected to return error")
			} else {
				assert.Equal(t, tc.expectedError.Error(), err.Error())
			}
		})
	}
}

func TestGenerateNodeClaimManifiest(t *testing.T) {
	t.Run("Should generate a machine object from the given workspace", func(t *testing.T) {
		mockWorkspace := utils.MockWorkspaceWithPreset

		machine := GenerateNodeClaimManifest(context.Background(), "0", mockWorkspace)

		assert.Check(t, machine != nil, "NodeClaim must not be nil")
		assert.Equal(t, machine.Namespace, mockWorkspace.Namespace, "NodeClaim must have same namespace as workspace")
	})
}
