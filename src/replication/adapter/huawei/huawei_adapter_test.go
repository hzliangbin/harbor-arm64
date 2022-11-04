package huawei

import (
	"os"
	"testing"

	adp "github.com/goharbor/harbor/src/replication/adapter"
	"github.com/goharbor/harbor/src/replication/model"
)

var hwAdapter adp.Adapter

func init() {
	var err error
	hwRegistry := &model.Registry{
		ID:          1,
		Name:        "Huawei",
		Description: "Adapter for SWR -- The image registry of Huawei Cloud",
		Type:        model.RegistryTypeHuawei,
		URL:         "https://swr.cn-north-1.myhuaweicloud.com",
		Credential:  &model.Credential{AccessKey: "cn-north-1@AQR6NF5G2MQ1V7U4FCD", AccessSecret: "2f7ec95070592fd4838a3aa4fd09338c047fd1cd654b3422197318f97281cd9"},
		Insecure:    false,
		Status:      "",
	}

	hwAdapter, err = newAdapter(hwRegistry)
	if err != nil {
		os.Exit(1)
	}

}

func TestAdapter_Info(t *testing.T) {
	info, err := hwAdapter.Info()
	if err != nil {
		t.Error(err)
	}
	t.Log(info)
}

func TestAdapter_PrepareForPush(t *testing.T) {
	repository := &model.Repository{
		Name:     "domain_repo_new",
		Metadata: make(map[string]interface{}),
	}
	resource := &model.Resource{}
	metadata := &model.ResourceMetadata{
		Repository: repository,
	}
	resource.Metadata = metadata
	err := hwAdapter.PrepareForPush([]*model.Resource{resource})
	if err != nil {
		t.Log("huawei ak/sk is not available", err.Error())
	} else {
		t.Log("success prepare for push")
	}
}

func TestAdapter_HealthCheck(t *testing.T) {
	health, err := hwAdapter.HealthCheck()
	if err != nil {
		t.Error(err)
	}
	t.Log(health)
}
