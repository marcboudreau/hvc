package hvc

import (
	"errors"
	"testing"
	"time"

	vault "github.com/hashicorp/vault/api"
	"github.com/marcboudreau/hvc/spec"
	"github.com/stretchr/testify/assert"
)

func TestNewCopy(t *testing.T) {
	for _, testcase := range []struct {
		spec        *spec.Copy
		sources     map[string]Vault
		errorAssert func(assert.TestingT, error, ...interface{}) bool
		copyAssert  func(assert.TestingT, interface{}, ...interface{}) bool
	}{
		// Happy path!
		{
			spec: &spec.Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*spec.CopyValue{
					"t1": {
						Source:     "s1",
						MountPoint: "kv",
						Path:       "where",
						Key:        "k1",
					},
				},
			},
			sources: map[string]Vault{
				"s1": &FakeVault{},
			},
			errorAssert: assert.NoError,
			copyAssert:  assert.NotNil,
		},
		// Error referencing non-existant source vault
		{
			spec: &spec.Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*spec.CopyValue{
					"t1": {
						Source:     "s1",
						MountPoint: "kv",
						Path:       "where",
						Key:        "k1",
					},
				},
			},
			errorAssert: assert.Error,
			copyAssert:  assert.Nil,
		},
	} {
		copy, err := NewCopy(testcase.spec, testcase.sources)
		testcase.errorAssert(t, err)
		testcase.copyAssert(t, copy)
	}
}

func TestNewCopySetsVaultSource(t *testing.T) {
	fakeVault := &FakeVault{
		name: "fake",
	}
	copy, err := NewCopy(&spec.Copy{
		MountPoint: "kv",
		Path:       "p1",
		Values: map[string]*spec.CopyValue{
			"t1": {
				Source:     "s1",
				MountPoint: "kv",
				Path:       "p1",
				Key:        "k1",
			},
		},
	}, map[string]Vault{
		"s1": fakeVault,
	})

	assert.NoError(t, err)
	assert.NotNil(t, copy)
	assert.Equal(t, fakeVault.Name(), copy.Values["t1"].Source.Name())
}

func TestNewCopyHandlesDefaultValues(t *testing.T) {
	for _, testcase := range []struct {
		spec                     *spec.Copy
		expectedMountPoint       string
		expectedValuesMountPoint string
		expectedValuesPath       string
		expectedValuesKey        string
	}{
		// mount-point omitted
		{
			spec: &spec.Copy{
				Path: "path1",
				Values: map[string]*spec.CopyValue{
					"k": {
						Source:     "s1",
						MountPoint: "vkv",
						Path:       "vpath1",
						Key:        "vk1",
					},
				},
			},
			expectedMountPoint:       "kv",
			expectedValuesMountPoint: "vkv",
			expectedValuesPath:       "vpath1",
			expectedValuesKey:        "vk1",
		},
		// values mount-point omitted
		{
			spec: &spec.Copy{
				MountPoint: "kv1",
				Path:       "path1",
				Values: map[string]*spec.CopyValue{
					"k": {
						Source: "s1",
						Path:   "vpath1",
						Key:    "vk1",
					},
				},
			},
			expectedMountPoint:       "kv1",
			expectedValuesMountPoint: "kv",
			expectedValuesPath:       "vpath1",
			expectedValuesKey:        "vk1",
		},
		// values path omitted
		{
			spec: &spec.Copy{
				MountPoint: "kv1",
				Path:       "path1",
				Values: map[string]*spec.CopyValue{
					"k": {
						Source:     "s1",
						MountPoint: "vkv1",
						Key:        "vk1",
					},
				},
			},
			expectedMountPoint:       "kv1",
			expectedValuesMountPoint: "vkv1",
			expectedValuesPath:       "path1",
			expectedValuesKey:        "vk1",
		},
		// values key omitted
		{
			spec: &spec.Copy{
				MountPoint: "kv1",
				Path:       "path1",
				Values: map[string]*spec.CopyValue{
					"k": {
						Source:     "s1",
						MountPoint: "vkv1",
						Path:       "vpath1",
					},
				},
			},
			expectedMountPoint:       "kv1",
			expectedValuesMountPoint: "vkv1",
			expectedValuesPath:       "vpath1",
			expectedValuesKey:        "k",
		},
	} {
		copy, _ := NewCopy(testcase.spec, map[string]Vault{"s1": &FakeVault{}})
		assert.Equal(t, testcase.expectedMountPoint, copy.MountPoint)
		assert.Equal(t, testcase.expectedValuesMountPoint, copy.Values["k"].MountPoint)
		assert.Equal(t, testcase.expectedValuesPath, copy.Values["k"].Path)
		assert.Equal(t, testcase.expectedValuesKey, copy.Values["k"].Key)
	}
}

func TestTargetUpdateTime(t *testing.T) {
	for _, testcase := range []struct {
		copy         *Copy
		vault        *FakeVault
		expectedTime time.Time
		errorAssert  func(assert.TestingT, error, ...interface{}) bool
	}{
		// Happy path
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
			},
			vault: &FakeVault{
				readResponses: []FakeVaultResponse{
					{
						secret: &vault.Secret{
							Data: map[string]interface{}{
								"updated_time": "2022-04-08T13:01:34.000000000Z",
							},
						},
						err: nil,
					},
				},
			},
			expectedTime: time.Date(2022, time.April, 8, 13, 01, 34, 0, time.UTC),
			errorAssert:  assert.NoError,
		},
		// Error Vault response
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
			},
			vault: &FakeVault{
				readResponses: []FakeVaultResponse{
					{
						secret: nil,
						err:    errors.New("error"),
					},
				},
			},
			expectedTime: time.Unix(0, 0),
			errorAssert:  assert.Error,
		},
		// Error parsing time in response
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
			},
			vault: &FakeVault{
				readResponses: []FakeVaultResponse{
					{
						secret: &vault.Secret{
							Data: map[string]interface{}{
								"updated_time": "bad",
							},
						},
						err: nil,
					},
				},
			},
			expectedTime: time.Unix(0, 0),
			errorAssert:  assert.Error,
		},
	} {
		time, err := testcase.copy.TargetUpdateTime(testcase.vault)
		testcase.errorAssert(t, err)
		assert.Equal(t, testcase.expectedTime, time)
	}
}

func TestDetermineNeedToCopy(t *testing.T) {
	for _, testcase := range []struct {
		copy           *Copy
		targetTime     time.Time
		expectedResult bool
		errorAssert    func(assert.TestingT, error, ...interface{}) bool
	}{
		// Happy path
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*CopyValue{
					"t1": {
						Source: &FakeVault{
							readResponses: []FakeVaultResponse{
								{
									secret: &vault.Secret{
										Data: map[string]interface{}{
											"updated_time": "2022-04-08T15:12:52.0000000000Z",
										},
									},
									err: nil,
								},
							},
						},
						MountPoint: "kv",
						Path:       "where",
						Key:        "k1",
					},
					"t2": {
						Source: &FakeVault{
							readResponses: []FakeVaultResponse{
								{
									secret: &vault.Secret{
										Data: map[string]interface{}{
											"updated_time": "2022-04-08T15:13:05.000000000Z",
										},
									},
									err: nil,
								},
							},
						},
						MountPoint: "kv",
						Path:       "other",
						Key:        "k1",
					},
					// "t3": {
					// 	Source:     &FakeVault{
					// 		readResponses: []FakeVaultResponse{
					// 			{
					// 				secret: &vault.Secret{
					// 					Data: map[string]interface{}{
					// 						"updated_time": "2022-04-08T15:13:46.000000000Z",
					// 					},
					// 				},
					// 				err: nil,
					// 			},
					// 		},
					// 	MountPoint: "keyvalue",
					// 	Path:       "where",
					// 	Key:        "k1",
					// },
				},
			},
			targetTime:     time.Date(2022, time.April, 8, 10, 0, 0, 0, time.UTC),
			expectedResult: true,
			errorAssert:    assert.NoError,
		},
		// Missing source secret
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*CopyValue{
					"t1": {
						Source: &FakeVault{
							readResponses: []FakeVaultResponse{
								{
									secret: nil,
									err:    errors.New("error"),
								},
							},
						},
						MountPoint: "secret",
						Path:       "where",
						Key:        "k1",
					},
				},
			},
			targetTime:  time.Unix(0, 0),
			errorAssert: assert.Error,
		},
	} {
		result, err := testcase.copy.DetermineNeedToCopy(testcase.targetTime)
		testcase.errorAssert(t, err)
		assert.Equal(t, testcase.expectedResult, result)
	}
}

func TestUpdateTargetSecret(t *testing.T) {
	for _, testcase := range []struct {
		copy        *Copy
		targetVault Vault
		errorAssert func(assert.TestingT, error, ...interface{}) bool
	}{
		// Happy path
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*CopyValue{
					"t1": {
						Source: &FakeVault{
							readResponses: []FakeVaultResponse{
								{
									secret: &vault.Secret{
										Data: map[string]interface{}{
											"data": map[string]interface{}{
												"k1": "ThePassword",
											},
										},
									},
									err: nil,
								},
							},
						},
						MountPoint: "kv",
						Path:       "where",
						Key:        "k1",
					},
				},
			},
			targetVault: &FakeVault{
				writeResponses: []FakeVaultResponse{
					{
						secret: &vault.Secret{},
						err:    nil,
					},
				},
			},
			errorAssert: assert.NoError,
		},
		// Error reading source value
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*CopyValue{
					"t1": {
						Source: &FakeVault{
							readResponses: []FakeVaultResponse{
								{
									secret: nil,
									err:    errors.New("error"),
								},
							},
						},
						MountPoint: "kv",
						Path:       "where",
						Key:        "k1",
					},
				},
			},
			targetVault: &FakeVault{},
			errorAssert: assert.Error,
		},
		// Error key missing from source secret
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*CopyValue{
					"t1": {
						Source: &FakeVault{
							readResponses: []FakeVaultResponse{
								{
									secret: &vault.Secret{
										Data: map[string]interface{}{
											"data": map[string]interface{}{},
										},
									},
									err: nil,
								},
							},
						},
						MountPoint: "kv",
						Path:       "where",
						Key:        "k1",
					},
				},
			},
			targetVault: &FakeVault{},
			errorAssert: assert.Error,
		},
		// Error writing target secret
		{
			copy: &Copy{
				MountPoint: "kv",
				Path:       "where",
				Values: map[string]*CopyValue{
					"t1": {
						Source: &FakeVault{
							readResponses: []FakeVaultResponse{
								{
									secret: &vault.Secret{
										Data: map[string]interface{}{
											"data": map[string]interface{}{
												"k1": "value",
											},
										},
									},
								},
							},
						},
						MountPoint: "kv",
						Path:       "where",
						Key:        "k1",
					},
				},
			},
			targetVault: &FakeVault{
				writeResponses: []FakeVaultResponse{
					{
						secret: nil,
						err:    errors.New("error"),
					},
				},
			},
			errorAssert: assert.Error,
		},
	} {
		testcase.errorAssert(t, testcase.copy.UpdateTargetSecret(testcase.targetVault))
	}
}
