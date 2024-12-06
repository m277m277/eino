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
//	mockgen -destination ../../internal/mock/components/model/ChatModel_mock.go --package model -source interface.go
//

// Package model is a generated GoMock package.
package model

import (
	context "context"
	reflect "reflect"

	model "github.com/cloudwego/eino/components/model"
	schema "github.com/cloudwego/eino/schema"
	gomock "go.uber.org/mock/gomock"
)

// MockChatModel is a mock of ChatModel interface.
type MockChatModel struct {
	ctrl     *gomock.Controller
	recorder *MockChatModelMockRecorder
}

// MockChatModelMockRecorder is the mock recorder for MockChatModel.
type MockChatModelMockRecorder struct {
	mock *MockChatModel
}

// NewMockChatModel creates a new mock instance.
func NewMockChatModel(ctrl *gomock.Controller) *MockChatModel {
	mock := &MockChatModel{ctrl: ctrl}
	mock.recorder = &MockChatModelMockRecorder{mock}
	return mock
}

// EXPECT returns an object that allows the caller to indicate expected use.
func (m *MockChatModel) EXPECT() *MockChatModelMockRecorder {
	return m.recorder
}

// BindTools mocks base method.
func (m *MockChatModel) BindTools(tools []*schema.ToolInfo) error {
	m.ctrl.T.Helper()
	ret := m.ctrl.Call(m, "BindTools", tools)
	ret0, _ := ret[0].(error)
	return ret0
}

// BindTools indicates an expected call of BindTools.
func (mr *MockChatModelMockRecorder) BindTools(tools any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "BindTools", reflect.TypeOf((*MockChatModel)(nil).BindTools), tools)
}

// Generate mocks base method.
func (m *MockChatModel) Generate(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, input}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Generate", varargs...)
	ret0, _ := ret[0].(*schema.Message)
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Generate indicates an expected call of Generate.
func (mr *MockChatModelMockRecorder) Generate(ctx, input any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, input}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Generate", reflect.TypeOf((*MockChatModel)(nil).Generate), varargs...)
}

// Stream mocks base method.
func (m *MockChatModel) Stream(ctx context.Context, input []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	m.ctrl.T.Helper()
	varargs := []any{ctx, input}
	for _, a := range opts {
		varargs = append(varargs, a)
	}
	ret := m.ctrl.Call(m, "Stream", varargs...)
	ret0, _ := ret[0].(*schema.StreamReader[*schema.Message])
	ret1, _ := ret[1].(error)
	return ret0, ret1
}

// Stream indicates an expected call of Stream.
func (mr *MockChatModelMockRecorder) Stream(ctx, input any, opts ...any) *gomock.Call {
	mr.mock.ctrl.T.Helper()
	varargs := append([]any{ctx, input}, opts...)
	return mr.mock.ctrl.RecordCallWithMethodType(mr.mock, "Stream", reflect.TypeOf((*MockChatModel)(nil).Stream), varargs...)
}
