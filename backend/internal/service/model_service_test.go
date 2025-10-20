package service_test

import (
	"context"
	"errors"
	"testing"

	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/llm/mocks"
	"flow-ai/backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func setupModelService(t *testing.T) (*service.ModelService, *mocks.MockLLMProvider) {
	mockLLMProvider := mocks.NewMockLLMProvider(t)
	modelService := service.NewModelService(mockLLMProvider)
	return modelService, mockLLMProvider
}

func TestModelService_List(t *testing.T) {
	ctx := context.Background()
	modelService, mockLLMProvider := setupModelService(t)

	expectedResponse := &llm.ListModelsResponse{
		Models: []llm.Model{{Name: "test-model"}},
	}
	expectedError := errors.New("provider error")

	testCases := []struct {
		name         string
		setupMock    func()
		expectError  bool
		expectedResp *llm.ListModelsResponse
		expectedErr  error
	}{
		{
			name: "Success",
			setupMock: func() {

				mockLLMProvider.On("ListModels", ctx).Return(expectedResponse, nil).Once()
			},
			expectError:  false,
			expectedResp: expectedResponse,
		},
		{
			name: "Failure - Provider Error",
			setupMock: func() {

				mockLLMProvider.On("ListModels", ctx).Return(nil, expectedError).Once()
			},
			expectError: true,
			expectedErr: expectedError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			resp, err := modelService.List(ctx)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}

			mockLLMProvider.AssertExpectations(t)
		})
	}
}

func TestModelService_Delete(t *testing.T) {
	ctx := context.Background()
	modelService, mockLLMProvider := setupModelService(t)

	req := &llm.DeleteModelRequest{Name: "test-model"}
	expectedError := errors.New("provider error")

	testCases := []struct {
		name        string
		setupMock   func()
		expectError bool
		expectedErr error
	}{
		{
			name: "Success",
			setupMock: func() {
				mockLLMProvider.On("DeleteModel", ctx, req).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Failure - Provider Error",
			setupMock: func() {
				mockLLMProvider.On("DeleteModel", ctx, req).Return(expectedError).Once()
			},
			expectError: true,
			expectedErr: expectedError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			err := modelService.Delete(ctx, req)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
			mockLLMProvider.AssertExpectations(t)
		})
	}
}

func TestModelService_Show(t *testing.T) {
	ctx := context.Background()
	modelService, mockLLMProvider := setupModelService(t)

	req := &llm.ShowModelRequest{Name: "test-model"}
	expectedResponse := &llm.ModelInfo{Modelfile: "FROM scratch"}
	expectedError := errors.New("provider error")

	testCases := []struct {
		name         string
		setupMock    func()
		expectError  bool
		expectedResp *llm.ModelInfo
		expectedErr  error
	}{
		{
			name: "Success",
			setupMock: func() {
				mockLLMProvider.On("ShowModelInfo", ctx, req).Return(expectedResponse, nil).Once()
			},
			expectError:  false,
			expectedResp: expectedResponse,
		},
		{
			name: "Failure - Provider Error",
			setupMock: func() {
				mockLLMProvider.On("ShowModelInfo", ctx, req).Return(nil, expectedError).Once()
			},
			expectError: true,
			expectedErr: expectedError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			resp, err := modelService.Show(ctx, req)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}
			mockLLMProvider.AssertExpectations(t)
		})
	}
}

func TestModelService_Pull(t *testing.T) {
	ctx := context.Background()
	modelService, mockLLMProvider := setupModelService(t)

	req := &llm.PullModelRequest{Name: "test-model"}
	expectedError := errors.New("provider error")

	testCases := []struct {
		name        string
		setupMock   func()
		expectError bool
		expectedErr error
	}{
		{
			name: "Success",
			setupMock: func() {

				mockLLMProvider.On("PullModel", ctx, req, mock.Anything).Return(nil).Once()
			},
			expectError: false,
		},
		{
			name: "Failure - Provider Error",
			setupMock: func() {
				mockLLMProvider.On("PullModel", ctx, req, mock.Anything).Return(expectedError).Once()
			},
			expectError: true,
			expectedErr: expectedError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			tc.setupMock()

			testChan := make(chan llm.PullStatus, 1)

			go func() {
				for range testChan {
				}
			}()

			err := modelService.Pull(ctx, req, testChan)

			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
			} else {
				assert.NoError(t, err)
			}
			mockLLMProvider.AssertExpectations(t)

			close(testChan)
		})
	}
}
