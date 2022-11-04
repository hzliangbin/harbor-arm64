// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package api

import (
	"fmt"
	"net/http"
	"testing"

	v1 "github.com/goharbor/harbor/src/pkg/scan/rest/v1"

	"github.com/goharbor/harbor/src/pkg/q"
	sc "github.com/goharbor/harbor/src/pkg/scan/api/scanner"
	"github.com/goharbor/harbor/src/pkg/scan/dao/scanner"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"github.com/stretchr/testify/suite"
)

const (
	rootRoute = "/api/scanners"
)

// ScannerAPITestSuite is test suite for testing the scanner API
type ScannerAPITestSuite struct {
	suite.Suite

	originC sc.Controller
	mockC   *MockScannerAPIController
}

// TestScannerAPI is the entry of ScannerAPITestSuite
func TestScannerAPI(t *testing.T) {
	suite.Run(t, new(ScannerAPITestSuite))
}

// SetupSuite prepares testing env
func (suite *ScannerAPITestSuite) SetupTest() {
	suite.originC = sc.DefaultController
	m := &MockScannerAPIController{}
	sc.DefaultController = m

	suite.mockC = m
}

// TearDownTest clears test case env
func (suite *ScannerAPITestSuite) TearDownTest() {
	// Restore
	sc.DefaultController = suite.originC
}

// TestScannerAPICreate tests the post request to create new one
func (suite *ScannerAPITestSuite) TestScannerAPIBase() {
	// Including general cases
	cases := []*codeCheckingCase{
		// 401
		{
			request: &testingRequest{
				url:    rootRoute,
				method: http.MethodPost,
			},
			code: http.StatusUnauthorized,
		},
		// 403
		{
			request: &testingRequest{
				url:        rootRoute,
				method:     http.MethodPost,
				credential: nonSysAdmin,
			},
			code: http.StatusForbidden,
		},
		// 400
		{
			request: &testingRequest{
				url:        rootRoute,
				method:     http.MethodPost,
				credential: sysAdmin,
				bodyJSON: &scanner.Registration{
					URL: "http://a.b.c",
				},
			},
			code: http.StatusBadRequest,
		},
	}

	runCodeCheckingCases(suite.T(), cases...)
}

// TestScannerAPIGet tests api get
func (suite *ScannerAPITestSuite) TestScannerAPIGet() {
	res := &scanner.Registration{
		ID:          1000,
		UUID:        "uuid",
		Name:        "TestScannerAPIGet",
		Description: "JUST FOR TEST",
		URL:         "https://a.b.c",
	}
	suite.mockC.On("GetRegistration", "uuid").Return(res, nil)

	// Get
	rr := &scanner.Registration{}
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("%s/%s", rootRoute, "uuid"),
		method:     http.MethodGet,
		credential: sysAdmin,
	}, rr)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), rr)
	assert.Equal(suite.T(), res.Name, rr.Name)
	assert.Equal(suite.T(), res.UUID, rr.UUID)
}

// TestScannerAPICreate tests create.
func (suite *ScannerAPITestSuite) TestScannerAPICreate() {
	r := &scanner.Registration{
		Name:        "TestScannerAPICreate",
		Description: "JUST FOR TEST",
		URL:         "https://a.b.c",
	}

	suite.mockQuery(r)
	suite.mockC.On("CreateRegistration", r).Return("uuid", nil)

	// Create
	res := make(map[string]string, 1)
	err := handleAndParse(
		&testingRequest{
			url:        rootRoute,
			method:     http.MethodPost,
			credential: sysAdmin,
			bodyJSON:   r,
		}, &res)
	require.NoError(suite.T(), err)
	require.Condition(suite.T(), func() (success bool) {
		success = res["uuid"] == "uuid"
		return
	})
}

// TestScannerAPIList tests list
func (suite *ScannerAPITestSuite) TestScannerAPIList() {
	query := &q.Query{
		PageNumber: 1,
		PageSize:   500,
	}
	ll := []*scanner.Registration{
		{
			ID:          1001,
			UUID:        "uuid",
			Name:        "TestScannerAPIList",
			Description: "JUST FOR TEST",
			URL:         "https://a.b.c",
		}}
	suite.mockC.On("ListRegistrations", query).Return(ll, nil)

	// List
	l := make([]*scanner.Registration, 0)
	err := handleAndParse(&testingRequest{
		url:        rootRoute,
		method:     http.MethodGet,
		credential: sysAdmin,
	}, &l)
	require.NoError(suite.T(), err)
	assert.Condition(suite.T(), func() (success bool) {
		success = len(l) > 0 && l[0].Name == ll[0].Name
		return
	})
}

// TestScannerAPIUpdate tests the update API
func (suite *ScannerAPITestSuite) TestScannerAPIUpdate() {
	before := &scanner.Registration{
		ID:          1002,
		UUID:        "uuid",
		Name:        "TestScannerAPIUpdate_before",
		Description: "JUST FOR TEST",
		URL:         "https://a.b.c",
	}

	updated := &scanner.Registration{
		ID:          1002,
		UUID:        "uuid",
		Name:        "TestScannerAPIUpdate",
		Description: "JUST FOR TEST",
		URL:         "https://a.b.c",
	}

	suite.mockQuery(updated)
	suite.mockC.On("UpdateRegistration", updated).Return(nil)
	suite.mockC.On("GetRegistration", "uuid").Return(before, nil)

	rr := &scanner.Registration{}
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("%s/%s", rootRoute, "uuid"),
		method:     http.MethodPut,
		credential: sysAdmin,
		bodyJSON:   updated,
	}, rr)
	require.NoError(suite.T(), err)
	require.NotNil(suite.T(), rr)

	assert.Equal(suite.T(), updated.Name, rr.Name)
	assert.Equal(suite.T(), updated.UUID, rr.UUID)
}

//
func (suite *ScannerAPITestSuite) TestScannerAPIDelete() {
	r := &scanner.Registration{
		ID:          1003,
		UUID:        "uuid",
		Name:        "TestScannerAPIDelete",
		Description: "JUST FOR TEST",
		URL:         "https://a.b.c",
	}

	suite.mockC.On("GetRegistration", "uuid").Return(r, nil)
	suite.mockC.On("DeleteRegistration", "uuid").Return(r, nil)

	deleted := &scanner.Registration{}
	err := handleAndParse(&testingRequest{
		url:        fmt.Sprintf("%s/%s", rootRoute, "uuid"),
		method:     http.MethodDelete,
		credential: sysAdmin,
	}, deleted)

	require.NoError(suite.T(), err)
	assert.Equal(suite.T(), r.UUID, deleted.UUID)
	assert.Equal(suite.T(), r.Name, deleted.Name)
}

// TestScannerAPISetDefault tests the set default
func (suite *ScannerAPITestSuite) TestScannerAPISetDefault() {
	suite.mockC.On("SetDefaultRegistration", "uuid").Return(nil)

	body := make(map[string]interface{}, 1)
	body["is_default"] = true
	runCodeCheckingCases(suite.T(), &codeCheckingCase{
		request: &testingRequest{
			url:        fmt.Sprintf("%s/%s", rootRoute, "uuid"),
			method:     http.MethodPatch,
			credential: sysAdmin,
			bodyJSON:   body,
		},
		code: http.StatusOK,
	})
}

func (suite *ScannerAPITestSuite) mockQuery(r *scanner.Registration) {
	kw := make(map[string]interface{}, 1)
	kw["name"] = r.Name
	query := &q.Query{
		Keywords: kw,
	}
	emptyL := make([]*scanner.Registration, 0)
	suite.mockC.On("ListRegistrations", query).Return(emptyL, nil)

	kw2 := make(map[string]interface{}, 1)
	kw2["url"] = r.URL
	query2 := &q.Query{
		Keywords: kw2,
	}
	suite.mockC.On("ListRegistrations", query2).Return(emptyL, nil)
}

// MockScannerAPIController is mock of scanner API controller
type MockScannerAPIController struct {
	mock.Mock
}

// ListRegistrations ...
func (m *MockScannerAPIController) ListRegistrations(query *q.Query) ([]*scanner.Registration, error) {
	args := m.Called(query)
	return args.Get(0).([]*scanner.Registration), args.Error(1)
}

// CreateRegistration ...
func (m *MockScannerAPIController) CreateRegistration(registration *scanner.Registration) (string, error) {
	args := m.Called(registration)
	return args.String(0), args.Error(1)
}

// GetRegistration ...
func (m *MockScannerAPIController) GetRegistration(registrationUUID string) (*scanner.Registration, error) {
	args := m.Called(registrationUUID)
	s := args.Get(0)
	if s == nil {
		return nil, args.Error(1)
	}

	return s.(*scanner.Registration), args.Error(1)
}

// RegistrationExists ...
func (m *MockScannerAPIController) RegistrationExists(registrationUUID string) bool {
	args := m.Called(registrationUUID)
	return args.Bool(0)
}

// UpdateRegistration ...
func (m *MockScannerAPIController) UpdateRegistration(registration *scanner.Registration) error {
	args := m.Called(registration)
	return args.Error(0)
}

// DeleteRegistration ...
func (m *MockScannerAPIController) DeleteRegistration(registrationUUID string) (*scanner.Registration, error) {
	args := m.Called(registrationUUID)
	s := args.Get(0)
	if s == nil {
		return nil, args.Error(1)
	}

	return s.(*scanner.Registration), args.Error(1)
}

// SetDefaultRegistration ...
func (m *MockScannerAPIController) SetDefaultRegistration(registrationUUID string) error {
	args := m.Called(registrationUUID)
	return args.Error(0)
}

// SetRegistrationByProject ...
func (m *MockScannerAPIController) SetRegistrationByProject(projectID int64, scannerID string) error {
	args := m.Called(projectID, scannerID)
	return args.Error(0)
}

// GetRegistrationByProject ...
func (m *MockScannerAPIController) GetRegistrationByProject(projectID int64) (*scanner.Registration, error) {
	args := m.Called(projectID)
	s := args.Get(0)
	if s == nil {
		return nil, args.Error(1)
	}

	return s.(*scanner.Registration), args.Error(1)
}

// Ping ...
func (m *MockScannerAPIController) Ping(registration *scanner.Registration) (*v1.ScannerAdapterMetadata, error) {
	args := m.Called(registration)
	sam := args.Get(0)
	if sam == nil {
		return nil, args.Error(1)
	}

	return sam.(*v1.ScannerAdapterMetadata), nil
}

// GetMetadata ...
func (m *MockScannerAPIController) GetMetadata(registrationUUID string) (*v1.ScannerAdapterMetadata, error) {
	args := m.Called(registrationUUID)
	sam := args.Get(0)
	if sam == nil {
		return nil, args.Error(1)
	}

	return sam.(*v1.ScannerAdapterMetadata), nil
}
