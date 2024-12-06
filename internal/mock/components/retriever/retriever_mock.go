/*
 * Copyright 2024 CloudWeGo Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

// Code generated by MockGen. DO NOT EDIT.
// Source: interface.go
//
// Generated by this command:
//
//	mockgen -destination ../../internal/mock/components/retriever/retriever_mock.go --package retriever -source interface.go
//

// Package retriever is a generated GoMock package.
package retriever

import (
	context "context"
	reflect "reflect"

	retriever "github.com/cloudwego/eino/components/retriever"
	schema "github.com/cloudwego/eino/schema"
	gomock "go.uber.org/mock/gomock"
)

// MockRetriever is a mock of Retriever interface.
type MockRetriever struct {
	ctrl     *gomock.Controller
	recorder *MockRetrieverMockRecorder
}

// MockRetrieverMockRecorder is the mock recorder for MockRetriever.
type MockRetrieverMockRecorder struct {
	mock *MockRetriever
}

// NewMockRetriever creates a new mock instance.
func NewMockRetriever(ctrl *gomock.Controller) *MockRetriever {
	mock := &MockRetriever{ctrl: ctrl}
	mock.recorder = &MockRetrieverMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockRetriever) EXPECT() *MockRetrieverMockRecorder {
	return m.recorder
}

// Retrieve mocks base method.
func (m *MockRetriever) Retrieve(ctx context.Context, query string, opts ...retriever.Option) ([]*schema.Document, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, query}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Retrieve", varargs...)
	ret0, _ := ret[0].([]*schema.Document)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Retrieve indicates an expected call of Retrieve.
func (mr *MockRetrieverMockRecorder) Retrieve(ctx, query any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, query}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Retrieve", reflect.TypeOf((*MockRetriever)(nil).Retrieve), varargs...)
}
