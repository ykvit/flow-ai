package service_test

import (
	"context"
	"errors"
	"testing"

	"flow-ai/backend/internal/llm"
	"flow-ai/backend/internal/llm/mocks" // Import the generated mock for LLMProvider
	"flow-ai/backend/internal/service"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

// setupModelService is a test helper that creates a ModelService with its
// LLMProvider dependency mocked.
//
// WHY: This pattern is a cornerstone of our unit testing strategy. It allows
// each test to get a fresh, isolated instance of the service and a controller
// for its dependency (`mockLLMProvider`), ensuring tests don't interfere with
// each other.
func setupModelService(t *testing.T) (*service.ModelService, *mocks.MockLLMProvider) {
	mockLLMProvider := mocks.NewMockLLMProvider(t)
	modelService := service.NewModelService(mockLLMProvider)
	return modelService, mockLLMProvider
}

// TestModelService_List uses a table-driven approach to test the `List` method.
//
// GOAL: Verify that the service correctly calls its dependency (`LLMProvider.ListModels`)
// and propagates both the successful response and any potential errors. Since
// `ModelService` is a thin wrapper, these tests confirm that the wiring is correct.
func TestModelService_List(t *testing.T) {
	// ARRANGE: Set up shared variables for the test cases.
	ctx := context.Background()
	modelService, mockLLMProvider := setupModelService(t)

	expectedResponse := &llm.ListModelsResponse{
		Models: []llm.Model{{Name: "test-model"}},
	}
	expectedError := errors.New("provider error")

	// Table-driven tests make it easy to test multiple scenarios for the same function.
	testCases := []struct {
		name         string // A descriptive name for the test case.
		setupMock    func() // A function to configure the mock's behavior for this specific case.
		expectError  bool   // Whether we expect an error to be returned.
		expectedResp *llm.ListModelsResponse
		expectedErr  error
	}{
		{
			name: "Success",
			setupMock: func() {
				// We configure the mock to expect a single call to `ListModels`.
				// If called, it should return our sample response and no error.
				mockLLMProvider.On("ListModels", ctx).Return(expectedResponse, nil).Once()
			},
			expectError:  false,
			expectedResp: expectedResponse,
		},
		{
			name: "Failure - Provider Error",
			setupMock: func() {
				// Here, we configure the mock to simulate a failure in the LLM provider.
				mockLLMProvider.On("ListModels", ctx).Return(nil, expectedError).Once()
			},
			expectError: true,
			expectedErr: expectedError,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// ARRANGE (for this sub-test)
			tc.setupMock()

			// ACT: Call the method under test.
			resp, err := modelService.List(ctx)

			// ASSERT: Check the results against our expectations.
			if tc.expectError {
				assert.Error(t, err)
				assert.Equal(t, tc.expectedErr, err)
				assert.Nil(t, resp)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tc.expectedResp, resp)
			}

			// This is a crucial step: it verifies that all expectations set on the
			// mock object (with `.On(...)`) were met. The test will fail if the
			// method was not called, or called with different arguments.
			mockLLMProvider.AssertExpectations(t)
		})
	}
}

// TestModelService_Delete follows the same table-driven pattern for the `Delete` method.
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

// TestModelService_Show follows the same table-driven pattern for the `Show` method.
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

// TestModelService_Pull tests the `Pull` method, which involves a channel.
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
				// For arguments that are complex or non-deterministic (like a channel),
				// `mock.Anything` is a useful matcher that accepts any value for that argument.
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

			// A goroutine is used to drain the channel. This prevents the test
			// from deadlocking if the code under test were to send data to it.
			go func() {
				for range testChan {
					// Discard any values received.
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

			// Clean up by closing the channel, which terminates the goroutine.
			close(testChan)
		})
	}
}
