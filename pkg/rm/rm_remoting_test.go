/*
 * Licensed to the Apache Software Foundation (ASF) under one or more
 * contributor license agreements.  See the NOTICE file distributed with
 * this work for additional information regarding copyright ownership.
 * The ASF licenses this file to You under the Apache License, Version 2.0
 * (the "License"); you may not use this file except in compliance with
 * the License.  You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package rm

import (
	"errors"
	"reflect"
	"testing"

	"github.com/agiledragon/gomonkey/v2"
	"github.com/golang/mock/gomock"
	"github.com/stretchr/testify/assert"

	"seata.apache.org/seata-go/pkg/protocol/branch"
	"seata.apache.org/seata-go/pkg/protocol/message"
	"seata.apache.org/seata-go/pkg/remoting/getty"
)

func TestGetRMRemotingInstance(t *testing.T) {
	tests := struct {
		name string
		want *RMRemoting
	}{"test1", &RMRemoting{}}

	t.Run(tests.name, func(t *testing.T) {
		assert.Equalf(t, tests.want, GetRMRemotingInstance(), "GetRMRemotingInstance()")
	})
}

func TestBranchRegister(t *testing.T) {
	r := &RMRemoting{}

	tests := []struct {
		name          string
		mockFunc      interface{}
		param         BranchRegisterParam
		wantBranchID  int64
		wantErrString string
	}{
		{
			name: "success",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.BranchRegisterResponse{
					AbstractTransactionResponse: message.AbstractTransactionResponse{
						AbstractResultMessage: message.AbstractResultMessage{
							ResultCode: message.ResultCodeSuccess,
						},
					},
					BranchId: 1001,
				}, nil
			},
			param:        BranchRegisterParam{Xid: "x1", LockKeys: "lk1", ResourceId: "r1", BranchType: 1},
			wantBranchID: 1001,
		},
		{
			name: "failed response",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.BranchRegisterResponse{
					AbstractTransactionResponse: message.AbstractTransactionResponse{
						AbstractResultMessage: message.AbstractResultMessage{
							ResultCode: message.ResultCodeFailed,
						},
					},
					BranchId: 1,
				}, nil
			},
			param:         BranchRegisterParam{Xid: "x1", LockKeys: "lk1", ResourceId: "r1", BranchType: 1},
			wantErrString: "Response",
		},
		{
			name: "send error",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return nil, errors.New("network error")
			},
			param:         BranchRegisterParam{Xid: "x1", LockKeys: "lk1", ResourceId: "r1", BranchType: 1},
			wantErrString: "network error",
			wantBranchID:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(getty.GetGettyRemotingClient()), "SendSyncRequest", tt.mockFunc)
			defer patch.Reset()

			id, err := r.BranchRegister(tt.param)
			if tt.wantErrString != "" {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErrString)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantBranchID, id)
			}
		})
	}
}

func TestBranchReport(t *testing.T) {
	r := &RMRemoting{}

	tests := []struct {
		name      string
		mockFunc  interface{}
		param     BranchReportParam
		wantErr   bool
		errString string
	}{
		{
			name: "success",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.BranchReportResponse{
					AbstractTransactionResponse: message.AbstractTransactionResponse{
						AbstractResultMessage: message.AbstractResultMessage{
							ResultCode: message.ResultCodeSuccess,
						},
					},
				}, nil
			},
			param:   BranchReportParam{Xid: "x1", BranchId: 1, Status: 1},
			wantErr: false,
		},
		{
			name: "failed result",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.BranchReportResponse{
					AbstractTransactionResponse: message.AbstractTransactionResponse{
						AbstractResultMessage: message.AbstractResultMessage{
							ResultCode: message.ResultCodeFailed,
						},
					},
				}, nil
			},
			param:   BranchReportParam{Xid: "x1", BranchId: 1, Status: 1},
			wantErr: true,
		},
		{
			name: "wrong response type",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return "wrong response type", nil
			},
			param:   BranchReportParam{Xid: "x1", BranchId: 1, Status: 1},
			wantErr: true,
		},
		{
			name: "send error",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return nil, errors.New("network error")
			},
			param:     BranchReportParam{Xid: "x1", BranchId: 1, Status: 1},
			wantErr:   true,
			errString: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(getty.GetGettyRemotingClient()), "SendSyncRequest", tt.mockFunc)
			defer patch.Reset()

			err := r.BranchReport(tt.param)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestLockQuery(t *testing.T) {
	r := &RMRemoting{}

	tests := []struct {
		name         string
		mockFunc     interface{}
		param        LockQueryParam
		wantLockable bool
		wantErr      bool
	}{
		{
			name: "lockable",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.GlobalLockQueryResponse{Lockable: true}, nil
			},
			param:        LockQueryParam{Xid: "x1", LockKeys: "lk1", ResourceId: "r1", BranchType: 1},
			wantLockable: true,
		},
		{
			name: "unlockable",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.GlobalLockQueryResponse{Lockable: false}, nil
			},
			param:        LockQueryParam{Xid: "x1", LockKeys: "lk1", ResourceId: "r1", BranchType: 1},
			wantLockable: false,
		},
		{
			name: "wrong response type",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return "wrong response type", nil
			},
			param:        LockQueryParam{Xid: "x1", LockKeys: "lk1", ResourceId: "r1", BranchType: 1},
			wantLockable: false,
		},
		{
			name: "send error",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return nil, errors.New("network error")
			},
			param:   LockQueryParam{Xid: "x1", LockKeys: "lk1", ResourceId: "r1", BranchType: 1},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(getty.GetGettyRemotingClient()), "SendSyncRequest", tt.mockFunc)
			defer patch.Reset()

			lockable, err := r.LockQuery(tt.param)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantLockable, lockable)
			}
		})
	}
}

func TestRegisterResource(t *testing.T) {
	r := &RMRemoting{}
	ctrl := gomock.NewController(t)
	defer ctrl.Finish()
	mockRes := NewMockResource(ctrl)
	mockRes.EXPECT().GetResourceId().Return("ospp").AnyTimes()
	mockRes.EXPECT().GetBranchType().Return(branch.BranchTypeAT).AnyTimes()
	mockRes.EXPECT().GetResourceGroupId().Return("").AnyTimes()

	tests := []struct {
		name      string
		mockFunc  interface{}
		resource  Resource
		wantErr   bool
		errString string
	}{
		{
			name: "register success",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.RegisterRMResponse{
					AbstractIdentifyResponse: message.AbstractIdentifyResponse{
						Identified: true,
					},
				}, nil
			},
			resource: mockRes,
			wantErr:  false,
		},
		{
			name: "register fail",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return message.RegisterRMResponse{
					AbstractIdentifyResponse: message.AbstractIdentifyResponse{
						Identified: false,
					},
				}, nil
			},
			resource: mockRes,
			wantErr:  false,
		},
		{
			name: "wrong response type",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return "wrong response type", nil
			},
			resource: mockRes,
			wantErr:  false,
		},
		{
			name: "send error",
			mockFunc: func(_ *getty.GettyRemotingClient, msg interface{}) (interface{}, error) {
				return nil, errors.New("network error")
			},
			resource:  mockRes,
			wantErr:   true,
			errString: "network error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			patch := gomonkey.ApplyMethod(reflect.TypeOf(getty.GetGettyRemotingClient()), "SendSyncRequest", tt.mockFunc)
			defer patch.Reset()

			err := r.RegisterResource(tt.resource)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errString)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
