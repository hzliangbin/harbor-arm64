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

package flow

import (
	"io"
	"os"
	"testing"

	"github.com/docker/distribution"
	"github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/config"
	"github.com/goharbor/harbor/src/replication/dao/models"
	"github.com/goharbor/harbor/src/replication/model"
	"github.com/goharbor/harbor/src/replication/operation/scheduler"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type fakedFactory struct {
}

func (fakedFactory) Create(*model.Registry) (adapter.Adapter, error) {
	return &fakedAdapter{}, nil
}

func (fakedFactory) AdapterPattern() *model.AdapterPattern {
	return nil
}

type fakedAdapter struct{}

func (f *fakedAdapter) Info() (*model.RegistryInfo, error) {
	return &model.RegistryInfo{
		Type: model.RegistryTypeHarbor,
		SupportedResourceTypes: []model.ResourceType{
			model.ResourceTypeImage,
			model.ResourceTypeChart,
		},
		SupportedTriggers: []model.TriggerType{model.TriggerTypeManual},
	}, nil
}

func (f *fakedAdapter) PrepareForPush([]*model.Resource) error {
	return nil
}
func (f *fakedAdapter) HealthCheck() (model.HealthStatus, error) {
	return model.Healthy, nil
}
func (f *fakedAdapter) FetchImages(filters []*model.Filter) ([]*model.Resource, error) {
	return []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}, nil
}

func (f *fakedAdapter) ManifestExist(repository, reference string) (exist bool, digest string, err error) {
	return false, "", nil
}
func (f *fakedAdapter) PullManifest(repository, reference string, accepttedMediaTypes []string) (manifest distribution.Manifest, digest string, err error) {
	return nil, "", nil
}
func (f *fakedAdapter) PushManifest(repository, reference, mediaType string, payload []byte) error {
	return nil
}
func (f *fakedAdapter) DeleteManifest(repository, digest string) error {
	return nil
}
func (f *fakedAdapter) BlobExist(repository, digest string) (exist bool, err error) {
	return false, nil
}
func (f *fakedAdapter) PullBlob(repository, digest string) (size int64, blob io.ReadCloser, err error) {
	return 0, nil, nil
}
func (f *fakedAdapter) PushBlob(repository, digest string, size int64, blob io.Reader) error {
	return nil
}
func (f *fakedAdapter) FetchCharts(filters []*model.Filter) ([]*model.Resource, error) {
	return []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/harbor",
				},
				Vtags: []string{"0.2.0"},
			},
		},
	}, nil
}
func (f *fakedAdapter) ChartExist(name, version string) (bool, error) {
	return false, nil
}
func (f *fakedAdapter) DownloadChart(name, version string) (io.ReadCloser, error) {
	return nil, nil
}
func (f *fakedAdapter) UploadChart(name, version string, chart io.Reader) error {
	return nil
}
func (f *fakedAdapter) DeleteChart(name, version string) error {
	return nil
}

type fakedScheduler struct{}

func (f *fakedScheduler) Preprocess(src []*model.Resource, dst []*model.Resource) ([]*scheduler.ScheduleItem, error) {
	items := []*scheduler.ScheduleItem{}
	for i, res := range src {
		items = append(items, &scheduler.ScheduleItem{
			SrcResource: res,
			DstResource: dst[i],
		})
	}
	return items, nil
}
func (f *fakedScheduler) Schedule(items []*scheduler.ScheduleItem) ([]*scheduler.ScheduleResult, error) {
	results := []*scheduler.ScheduleResult{}
	for _, item := range items {
		results = append(results, &scheduler.ScheduleResult{
			TaskID: item.TaskID,
			Error:  nil,
		})
	}
	return results, nil
}
func (f *fakedScheduler) Stop(id string) error {
	return nil
}

type fakedExecutionManager struct {
	taskID int64
}

func (f *fakedExecutionManager) Create(*models.Execution) (int64, error) {
	return 1, nil
}
func (f *fakedExecutionManager) List(...*models.ExecutionQuery) (int64, []*models.Execution, error) {
	return 0, nil, nil
}
func (f *fakedExecutionManager) Get(int64) (*models.Execution, error) {
	return &models.Execution{}, nil
}
func (f *fakedExecutionManager) Update(*models.Execution, ...string) error {
	return nil
}
func (f *fakedExecutionManager) Remove(int64) error {
	return nil
}
func (f *fakedExecutionManager) RemoveAll(int64) error {
	return nil
}
func (f *fakedExecutionManager) CreateTask(*models.Task) (int64, error) {
	f.taskID++
	id := f.taskID
	return id, nil
}
func (f *fakedExecutionManager) ListTasks(...*models.TaskQuery) (int64, []*models.Task, error) {
	return 0, nil, nil
}
func (f *fakedExecutionManager) GetTask(int64) (*models.Task, error) {
	return nil, nil
}
func (f *fakedExecutionManager) UpdateTask(*models.Task, ...string) error {
	return nil
}
func (f *fakedExecutionManager) UpdateTaskStatus(int64, string, int64, ...string) error {
	return nil
}
func (f *fakedExecutionManager) RemoveTask(int64) error {
	return nil
}
func (f *fakedExecutionManager) RemoveAllTasks(int64) error {
	return nil
}
func (f *fakedExecutionManager) GetTaskLog(int64) ([]byte, error) {
	return nil, nil
}

func TestMain(m *testing.M) {
	url := "https://registry.harbor.local"
	config.Config = &config.Configuration{
		CoreURL: url,
	}
	if err := adapter.RegisterFactory(model.RegistryTypeHarbor, new(fakedFactory)); err != nil {
		os.Exit(1)
	}
	os.Exit(m.Run())
}

func TestFetchResources(t *testing.T) {
	adapter := &fakedAdapter{}
	policy := &model.Policy{}
	resources, err := fetchResources(adapter, policy)
	require.Nil(t, err)
	assert.Equal(t, 2, len(resources))
}

func TestFilterResources(t *testing.T) {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeImage,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
				// TODO test labels
				Labels: nil,
			},
			Deleted:  true,
			Override: true,
		},
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/harbor",
				},
				Vtags: []string{"0.2.0", "0.3.0"},
				// TODO test labels
				Labels: nil,
			},
			Deleted:  true,
			Override: true,
		},
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/mysql",
				},
				Vtags: []string{"1.0"},
				// TODO test labels
				Labels: nil,
			},
			Deleted:  true,
			Override: true,
		},
	}
	filters := []*model.Filter{
		{
			Type:  model.FilterTypeResource,
			Value: model.ResourceTypeChart,
		},
		{
			Type:  model.FilterTypeName,
			Value: "library/*",
		},
		{
			Type:  model.FilterTypeName,
			Value: "library/harbor",
		},
		{
			Type:  model.FilterTypeTag,
			Value: "0.2.?",
		},
	}
	res, err := filterResources(resources, filters)
	require.Nil(t, err)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, "library/harbor", res[0].Metadata.Repository.Name)
	assert.Equal(t, 1, len(res[0].Metadata.Vtags))
	assert.Equal(t, "0.2.0", res[0].Metadata.Vtags[0])
}

func TestAssembleSourceResources(t *testing.T) {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	policy := &model.Policy{
		SrcRegistry: &model.Registry{
			ID: 1,
		},
	}
	res := assembleSourceResources(resources, policy)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, int64(1), res[0].Registry.ID)
}

func TestAssembleDestinationResources(t *testing.T) {
	resources := []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	policy := &model.Policy{
		DestRegistry:  &model.Registry{},
		DestNamespace: "test",
		Override:      true,
	}
	res := assembleDestinationResources(resources, policy)
	assert.Equal(t, 1, len(res))
	assert.Equal(t, model.ResourceTypeChart, res[0].Type)
	assert.Equal(t, "test/hello-world", res[0].Metadata.Repository.Name)
	assert.Equal(t, 1, len(res[0].Metadata.Vtags))
	assert.Equal(t, "latest", res[0].Metadata.Vtags[0])
}

func TestPreprocess(t *testing.T) {
	scheduler := &fakedScheduler{}
	srcResources := []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "library/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	dstResources := []*model.Resource{
		{
			Type: model.ResourceTypeChart,
			Metadata: &model.ResourceMetadata{
				Repository: &model.Repository{
					Name: "test/hello-world",
				},
				Vtags: []string{"latest"},
			},
			Override: false,
		},
	}
	items, err := preprocess(scheduler, srcResources, dstResources)
	require.Nil(t, err)
	assert.Equal(t, 1, len(items))
}

func TestCreateTasks(t *testing.T) {
	mgr := &fakedExecutionManager{}
	items := []*scheduler.ScheduleItem{
		{
			SrcResource: &model.Resource{},
			DstResource: &model.Resource{},
		},
	}
	err := createTasks(mgr, 1, items)
	require.Nil(t, err)
	assert.Equal(t, int64(1), items[0].TaskID)
}

func TestSchedule(t *testing.T) {
	sched := &fakedScheduler{}
	mgr := &fakedExecutionManager{}
	items := []*scheduler.ScheduleItem{
		{
			SrcResource: &model.Resource{},
			DstResource: &model.Resource{},
			TaskID:      1,
		},
	}
	n, err := schedule(sched, mgr, items)
	require.Nil(t, err)
	assert.Equal(t, 1, n)
}

func TestReplaceNamespace(t *testing.T) {
	// empty namespace
	repository := "c"
	namespace := ""
	result := replaceNamespace(repository, namespace)
	assert.Equal(t, "c", result)
	// repository contains no "/"
	repository = "c"
	namespace = "n"
	result = replaceNamespace(repository, namespace)
	assert.Equal(t, "n/c", result)
	// repository contains only one "/"
	repository = "b/c"
	namespace = "n"
	result = replaceNamespace(repository, namespace)
	assert.Equal(t, "n/c", result)
	// repository contains more than one "/"
	repository = "a/b/c"
	namespace = "n"
	result = replaceNamespace(repository, namespace)
	assert.Equal(t, "n/c", result)
}
