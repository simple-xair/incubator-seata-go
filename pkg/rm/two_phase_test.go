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
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"

	"seata.apache.org/seata-go/pkg/tm"
	testdata2 "seata.apache.org/seata-go/testdata"
)

var UserProviderInstance = NewTwoPhaseDemoService()

type UserProvider struct {
	Prepare       func(ctx context.Context, params ...interface{}) (bool, error)                           `seataTwoPhaseAction:"prepare" seataTwoPhaseServiceName:"TwoPhaseDemoService"`
	Commit        func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"commit"`
	Rollback      func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"rollback"`
	GetActionName func() string
}

func NewTwoPhaseDemoService() *UserProvider {
	return &UserProvider{
		Prepare: func(ctx context.Context, params ...interface{}) (bool, error) {
			return false, fmt.Errorf("execute two phase prepare method, param %v", params)
		},
		Commit: func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) {
			return false, fmt.Errorf("execute two phase commit method, xid %v", businessActionContext.Xid)
		},
		Rollback: func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) {
			return true, nil
		},
		GetActionName: func() string {
			return "TwoPhaseDemoService"
		},
	}
}

func TestParseTwoPhaseActionGetMethodName(t *testing.T) {
	tests := []struct {
		service      interface{}
		wantService  *TwoPhaseAction
		wantHasError bool
		wantErrMsg   string
	}{{
		service: &struct {
			Prepare111 func(ctx context.Context, params string) (bool, error)                                   `seataTwoPhaseAction:"prepare"`
			Commit111  func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"commit"`
			Rollback11 func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"rollback" seataTwoPhaseServiceName:"seataTwoPhaseName"`
			GetName    func() string
		}{},
		wantService: &TwoPhaseAction{
			actionName:         "seataTwoPhaseName",
			prepareMethodName:  "Prepare111",
			commitMethodName:   "Commit111",
			rollbackMethodName: "Rollback11",
		},
		wantHasError: false,
	}, {
		service: &struct {
			Prepare  func(ctx context.Context, params interface{}) (bool, error)                              `seataTwoPhaseAction:"prepare"`
			Commit   func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"commit"`
			Rollback func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"rollback"`
			GetName  func() string
		}{},
		wantService:  nil,
		wantHasError: true,
		wantErrMsg:   "missing two phase name",
	}, {
		service: &struct {
			Prepare  func(ctx context.Context, params interface{}) (bool, error)                              `serviceName:"seataTwoPhaseName"`
			Commit   func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"commit"`
			Rollback func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"rollback"`
			GetName  func() string
		}{},
		wantService:  nil,
		wantHasError: true,
		wantErrMsg:   "missing prepare method",
	}, {
		service: &struct {
			Prepare  func(ctx context.Context, params interface{}) (bool, error) `seataTwoPhaseAction:"commit" seataTwoPhaseName:"seataTwoPhaseName"`
			Commit   func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error)
			Rollback func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"rollback"`
			GetName  func() string
		}{},
		wantService:  nil,
		wantHasError: true,
		wantErrMsg:   "missing prepare method",
	}, {
		service: &struct {
			Prepare  func(ctx context.Context, params interface{}) (bool, error)                              `seataTwoPhaseAction:"prepare" seataTwoPhaseName:"seataTwoPhaseName"`
			Commit   func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"commit"`
			Rollback func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error)
			GetName  func() string
		}{},
		wantService:  nil,
		wantHasError: true,
		wantErrMsg:   "missing rollback method",
	}, {
		service: &struct {
			Prepare  func(ctx context.Context, params interface{}) (bool, error)                              `seataTwoPhaseAction:"prepare"`
			Commit   func(ctx context.Context, businessActionContext tm.BusinessActionContext) (bool, error)  `seataTwoPhaseAction:"commit"`
			Rollback func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"rollback" seataTwoPhaseServiceName:"seataTwoPhaseName"`
			GetName  func() string
		}{},
		wantService:  nil,
		wantHasError: true,
		wantErrMsg:   "missing commit method",
	}, {
		service: &struct {
			Prepare  func(ctx context.Context, params interface{}) (bool, error)                              `seataTwoPhaseAction:"prepare"`
			Commit   func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"commit"`
			Rollback func(ctx context.Context, businessActionContext tm.BusinessActionContext) (bool, error)  `seataTwoPhaseAction:"rollback" seataTwoPhaseServiceName:"seataTwoPhaseName"`
			GetName  func() string
		}{},
		wantService:  nil,
		wantHasError: true,
		wantErrMsg:   "missing rollback method",
	}}

	for _, tt := range tests {
		actual, err := ParseTwoPhaseAction(tt.service)
		if tt.wantHasError {
			assert.NotNil(t, err)
			assert.Equal(t, tt.wantErrMsg, err.Error())
		} else {
			assert.Nil(t, err)
			assert.Equal(t, actual.actionName, tt.wantService.actionName)
			assert.Equal(t, actual.prepareMethodName, tt.wantService.prepareMethodName)
			assert.Equal(t, actual.commitMethodName, tt.wantService.commitMethodName)
			assert.Equal(t, actual.rollbackMethodName, tt.wantService.rollbackMethodName)
		}
	}
}

type TwoPhaseDemoService1 struct {
	TwoPhasePrepare     func(ctx context.Context, params interface{}) (bool, error)                              `seataTwoPhaseAction:"prepare" seataTwoPhaseServiceName:"twoPhaseDemoService"`
	TwoPhaseCommit      func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"commit"`
	TwoPhaseRollback    func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) `seataTwoPhaseAction:"rollback"`
	TwoPhaseDemoService func() string
}

func NewTwoPhaseDemoService1() *TwoPhaseDemoService1 {
	return &TwoPhaseDemoService1{
		TwoPhasePrepare: func(ctx context.Context, params interface{}) (bool, error) {
			return false, fmt.Errorf("execute two phase prepare method, param %v", params)
		},
		TwoPhaseCommit: func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) {
			return false, fmt.Errorf("execute two phase commit method, xid %v", businessActionContext.Xid)
		},
		TwoPhaseRollback: func(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) {
			return true, nil
		},
		TwoPhaseDemoService: func() string {
			return "twoPhaseDemoService"
		},
	}
}

func TestParseTwoPhaseActionExecuteMethod1(t *testing.T) {
	twoPhaseService, err := ParseTwoPhaseAction(NewTwoPhaseDemoService1())
	ctx := context.Background()
	assert.Nil(t, err)
	assert.Equal(t, "TwoPhasePrepare", twoPhaseService.prepareMethodName)
	assert.Equal(t, "TwoPhaseCommit", twoPhaseService.commitMethodName)
	assert.Equal(t, "TwoPhaseRollback", twoPhaseService.rollbackMethodName)
	assert.Equal(t, "twoPhaseDemoService", twoPhaseService.actionName)

	resp, err := twoPhaseService.Prepare(ctx, 11)
	assert.Equal(t, false, resp)
	assert.Equal(t, "execute two phase prepare method, param 11", err.Error())

	resp, err = twoPhaseService.Commit(ctx, &tm.BusinessActionContext{Xid: "1234"})
	assert.Equal(t, false, resp)
	assert.Equal(t, "execute two phase commit method, xid 1234", err.Error())

	resp, err = twoPhaseService.Rollback(ctx, &tm.BusinessActionContext{Xid: "1234"})
	assert.Equal(t, true, resp)
	assert.Nil(t, err)

	assert.Equal(t, "twoPhaseDemoService", twoPhaseService.GetActionName())
}

type TwoPhaseDemoService2 struct{}

func (t *TwoPhaseDemoService2) Prepare(ctx context.Context, params interface{}) (bool, error) {
	return false, fmt.Errorf("execute two phase prepare method, param %v", params)
}

func (t *TwoPhaseDemoService2) Commit(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) {
	return true, fmt.Errorf("execute two phase commit method, xid %v", businessActionContext.Xid)
}

func (t *TwoPhaseDemoService2) Rollback(ctx context.Context, businessActionContext *tm.BusinessActionContext) (bool, error) {
	return false, fmt.Errorf("execute two phase rollback method, xid %v", businessActionContext.Xid)
}

func (t *TwoPhaseDemoService2) GetActionName() string {
	return "TestTwoPhaseDemoService"
}

func TestParseTwoPhaseActionExecuteMethod2(t *testing.T) {
	twoPhaseService, err := ParseTwoPhaseAction(&TwoPhaseDemoService2{})
	ctx := context.Background()
	assert.Nil(t, err)
	resp, err := twoPhaseService.Prepare(ctx, 11)
	assert.Equal(t, false, resp)
	assert.Equal(t, "execute two phase prepare method, param 11", err.Error())

	resp, err = twoPhaseService.Commit(ctx, &tm.BusinessActionContext{Xid: "1234"})
	assert.Equal(t, true, resp)
	assert.Equal(t, "execute two phase commit method, xid 1234", err.Error())

	resp, err = twoPhaseService.Rollback(ctx, &tm.BusinessActionContext{Xid: "1234"})
	assert.Equal(t, false, resp)
	assert.Equal(t, "execute two phase rollback method, xid 1234", err.Error())
}

func TestIsTwoPhaseAction(t *testing.T) {
	userProvider := &testdata2.TestTwoPhaseService{}
	userProvider1 := UserProviderInstance
	type args struct {
		v interface{}
	}

	tests := []struct {
		name string
		args args
		want bool
	}{
		{"test1", args{v: userProvider}, true},
		{"test2", args{v: userProvider1}, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equalf(t, tt.want, IsTwoPhaseAction(tt.args.v), "IsTwoPhaseAction(%v)", tt.args.v)
		})
	}
}

func TestParseTwoPhaseAction(t *testing.T) {
	type args struct {
		v interface{}
	}

	userProvider := UserProviderInstance
	twoPhaseAction, _ := ParseTwoPhaseAction(userProvider)
	args1 := args{v: userProvider}

	tests := struct {
		name    string
		args    args
		want    *TwoPhaseAction
		wantErr assert.ErrorAssertionFunc
	}{"test1", args1, twoPhaseAction, assert.NoError}

	t.Run(tests.name, func(t *testing.T) {
		got, err := ParseTwoPhaseAction(tests.args.v)
		if !tests.wantErr(t, err, fmt.Sprintf("ParseTwoPhaseAction(%v)", tests.args.v)) {
			return
		}
		assert.Equalf(t, tests.want.GetTwoPhaseService(), got.GetTwoPhaseService(), "ParseTwoPhaseAction(%v)", tests.args.v)
	})
}

func TestParseTwoPhaseActionByInterface(t *testing.T) {
	type args struct {
		v interface{}
	}

	userProvider := &UserProvider{}
	validAction, _ := ParseTwoPhaseAction(userProvider)

	var notStruct int = 10

	tests := []struct {
		name       string
		args       args
		want       *TwoPhaseAction
		wantErr    bool
		wantErrMsg string
	}{
		{
			name:    "success_struct",
			args:    args{v: userProvider},
			want:    validAction,
			wantErr: false,
		},
		{
			name:       "error_not_struct",
			args:       args{v: &notStruct},
			want:       nil,
			wantErr:    true,
			wantErrMsg: "param should be a struct, instead of a pointer",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParseTwoPhaseActionByInterface(tt.args.v)

			if tt.wantErr {
				assert.Nil(t, got)
				assert.Error(t, err)
				assert.Equal(t, tt.wantErrMsg, err.Error())
				return
			}

			assert.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestGetActionsMethodNames(t *testing.T) {
	type args struct {
		v interface{}
	}

	userProvider := NewTwoPhaseDemoService1()
	args1 := args{v: userProvider}

	tests := struct {
		name    string
		args    args
		wantErr assert.ErrorAssertionFunc
	}{
		"test1",
		args1,
		assert.NoError,
	}

	t.Run(tests.name, func(t *testing.T) {
		got, err := ParseTwoPhaseAction(tests.args.v)
		if !tests.wantErr(t, err) {
			return
		}

		assert.Equal(t, "TwoPhasePrepare", got.GetPrepareMethodName())
		assert.Equal(t, "TwoPhaseCommit", got.GetCommitMethodName())
		assert.Equal(t, "TwoPhaseRollback", got.GetRollbackMethodName())
	})
}

func TestGetPrepareAction(t *testing.T) {
	type testStruct struct {
		NotFunc        string
		WrongReturn0   func(ctx context.Context, params interface{}) (string, error) `seataTwoPhaseAction:"prepare"`
		WrongReturn1   func(ctx context.Context, params interface{}) (bool, string)  `seataTwoPhaseAction:"prepare"`
		NoParams       func() (bool, error)                                          `seataTwoPhaseAction:"prepare"`
		WrongNumOut    func(ctx context.Context) bool                                `seataTwoPhaseAction:"prepare"`
		WrongFirstType func(ctx string, params interface{}) (bool, error)            `seataTwoPhaseAction:"prepare"`
		ValidPrepare   func(ctx context.Context, params interface{}) (bool, error)   `seataTwoPhaseAction:"prepare"`
	}

	val := reflect.ValueOf(&testStruct{}).Elem()
	typ := val.Type()
	invalidValue := reflect.Value{}

	getFields := func(name string) reflect.StructField {
		field, _ := typ.FieldByName(name)
		return field
	}

	tests := []struct {
		name   string
		field  reflect.StructField
		value  reflect.Value
		wantOk bool
	}{
		{"notFunc", getFields("NotFunc"), val.FieldByName("NotFunc"), false},
		{"wrongReturn0", getFields("WrongReturn0"), val.FieldByName("WrongReturn0"), false},
		{"wrongReturn1", getFields("WrongReturn1"), val.FieldByName("WrongReturn1"), false},
		{"noParams", getFields("NoParams"), val.FieldByName("NoParams"), false},
		{"wrongNumOut", getFields("WrongNumOut"), val.FieldByName("WrongNumOut"), false},
		{"wrongFirstType", getFields("WrongFirstType"), val.FieldByName("WrongFirstType"), false},
		{"validPrepare", getFields("ValidPrepare"), val.FieldByName("ValidPrepare"), true},
		{"invalidValue", getFields("ValidPrepare"), invalidValue, false}, // !f.IsValid()
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, _, ok := getPrepareAction(tt.field, tt.value)
			if tt.wantOk {
				assert.True(t, ok)
				assert.Equal(t, tt.field.Name, name)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestGetCommitMethod(t *testing.T) {
	type testStruct struct {
		NotFunc        string
		WrongReturn0   func(ctx context.Context, bac *tm.BusinessActionContext) (string, error) `seataTwoPhaseAction:"commit"`
		WrongReturn1   func(ctx context.Context, bac *tm.BusinessActionContext) (bool, string)  `seataTwoPhaseAction:"commit"`
		NoParams       func() (bool, error)                                                     `seataTwoPhaseAction:"commit"`
		WrongNumOut    func(ctx context.Context) bool                                           `seataTwoPhaseAction:"commit"`
		WrongFirstType func(ctx string, bac *tm.BusinessActionContext) (bool, error)            `seataTwoPhaseAction:"commit"`
		ValidCommit    func(ctx context.Context, bac *tm.BusinessActionContext) (bool, error)   `seataTwoPhaseAction:"commit"`
	}

	val := reflect.ValueOf(&testStruct{}).Elem()
	typ := val.Type()
	invalidValue := reflect.Value{}

	getFields := func(name string) reflect.StructField {
		field, _ := typ.FieldByName(name)
		return field
	}

	tests := []struct {
		name   string
		field  reflect.StructField
		value  reflect.Value
		wantOk bool
	}{
		{"notFunc", getFields("NotFunc"), val.FieldByName("NotFunc"), false},
		{"wrongReturn0", getFields("WrongReturn0"), val.FieldByName("WrongReturn0"), false},
		{"wrongReturn1", getFields("WrongReturn1"), val.FieldByName("WrongReturn1"), false},
		{"noParams", getFields("NoParams"), val.FieldByName("NoParams"), false},
		{"wrongNumOut", getFields("WrongNumOut"), val.FieldByName("WrongNumOut"), false},
		{"wrongFirstType", getFields("WrongFirstType"), val.FieldByName("WrongFirstType"), false},
		{"validCommit", getFields("ValidCommit"), val.FieldByName("ValidCommit"), true},
		{"invalidValue", getFields("ValidCommit"), invalidValue, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, _, ok := getCommitMethod(tt.field, tt.value)
			if tt.wantOk {
				assert.True(t, ok)
				assert.Equal(t, tt.field.Name, name)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestGetRollbackMethod(t *testing.T) {
	type testStruct struct {
		NotFunc        string
		WrongReturn0   func(ctx context.Context, bac *tm.BusinessActionContext) (string, error) `seataTwoPhaseAction:"rollback"`
		WrongReturn1   func(ctx context.Context, bac *tm.BusinessActionContext) (bool, string)  `seataTwoPhaseAction:"rollback"`
		NoParams       func() (bool, error)                                                     `seataTwoPhaseAction:"rollback"`
		WrongNumOut    func(ctx context.Context) bool                                           `seataTwoPhaseAction:"rollback"`
		WrongFirstType func(ctx string, bac *tm.BusinessActionContext) (bool, error)            `seataTwoPhaseAction:"rollback"`
		ValidRollback  func(ctx context.Context, bac *tm.BusinessActionContext) (bool, error)   `seataTwoPhaseAction:"rollback"`
	}

	val := reflect.ValueOf(&testStruct{}).Elem()
	typ := val.Type()
	invalidValue := reflect.Value{}

	getFields := func(name string) reflect.StructField {
		field, _ := typ.FieldByName(name)
		return field
	}

	tests := []struct {
		name   string
		field  reflect.StructField
		value  reflect.Value
		wantOk bool
	}{
		{"notFunc", getFields("NotFunc"), val.FieldByName("NotFunc"), false},
		{"wrongReturn0", getFields("WrongReturn0"), val.FieldByName("WrongReturn0"), false},
		{"wrongReturn1", getFields("WrongReturn1"), val.FieldByName("WrongReturn1"), false},
		{"wrongNumOut", getFields("WrongNumOut"), val.FieldByName("WrongNumOut"), false},
		{"noParams", getFields("NoParams"), val.FieldByName("NoParams"), false},
		{"wrongFirstType", getFields("WrongFirstType"), val.FieldByName("WrongFirstType"), false},
		{"validRollback", getFields("ValidRollback"), val.FieldByName("ValidRollback"), true},
		{"invalidValue", getFields("ValidRollback"), invalidValue, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			name, _, ok := getRollbackMethod(tt.field, tt.value)
			if tt.wantOk {
				assert.True(t, ok)
				assert.Equal(t, tt.field.Name, name)
			} else {
				assert.False(t, ok)
			}
		})
	}
}

func TestGetActionName(t *testing.T) {
	type WithTag struct {
		Field1 int `seataTwoPhaseServiceName:"MyAction"`
		Field2 int
	}

	type WithoutTag struct {
		Field1 int
		Field2 int
	}

	type EmptyStruct struct{}

	tests := []struct {
		name string
		arg  interface{}
		want string
	}{
		{
			name: "struct with tag",
			arg:  &WithTag{},
			want: "MyAction",
		},
		{
			name: "struct without tag",
			arg:  &WithoutTag{},
			want: "",
		},
		{
			name: "empty struct pointer",
			arg:  &EmptyStruct{},
			want: "",
		},
		{
			name: "non-struct pointer",
			arg:  new(int),
			want: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getActionName(tt.arg)
			assert.Equal(t, tt.want, got)
		})
	}
}
