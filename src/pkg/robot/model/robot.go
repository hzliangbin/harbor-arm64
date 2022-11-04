package model

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/beego/beego/orm"
	"github.com/beego/beego/validation"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
)

// RobotTable is the name of table in DB that holds the robot object
const RobotTable = "robot"

func init() {
	orm.RegisterModel(&Robot{})
}

// Robot holds the details of a robot.
type Robot struct {
	ID           int64     `orm:"pk;auto;column(id)" json:"id"`
	Name         string    `orm:"column(name)" json:"name"`
	Token        string    `orm:"-" json:"token"`
	Description  string    `orm:"column(description)" json:"description"`
	ProjectID    int64     `orm:"column(project_id)" json:"project_id"`
	ExpiresAt    int64     `orm:"column(expiresat)" json:"expires_at"`
	Disabled     bool      `orm:"column(disabled)" json:"disabled"`
	Visible      bool      `orm:"column(visible)" json:"-"`
	CreationTime time.Time `orm:"column(creation_time);auto_now_add" json:"creation_time"`
	UpdateTime   time.Time `orm:"column(update_time);auto_now" json:"update_time"`
}

// TableName ...
func (r *Robot) TableName() string {
	return RobotTable
}

// FromJSON parses robot from json data
func (r *Robot) FromJSON(jsonData string) error {
	if len(jsonData) == 0 {
		return errors.New("empty json data to parse")
	}

	return json.Unmarshal([]byte(jsonData), r)
}

// ToJSON marshals Robot to JSON data
func (r *Robot) ToJSON() (string, error) {
	data, err := json.Marshal(r)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// RobotQuery ...
type RobotQuery struct {
	Name           string
	ProjectID      int64
	Disabled       bool
	FuzzyMatchName bool
	Pagination
}

// RobotCreate ...
type RobotCreate struct {
	Name        string         `json:"name"`
	ProjectID   int64          `json:"pid"`
	Description string         `json:"description"`
	Disabled    bool           `json:"disabled"`
	Visible     bool           `json:"-"`
	Access      []*rbac.Policy `json:"access"`
}

// Pagination ...
type Pagination struct {
	Page int64
	Size int64
}

// Valid ...
func (rq *RobotCreate) Valid(v *validation.Validation) {
	if utils.IsIllegalLength(rq.Name, 1, 255) {
		v.SetError("name", "robot name with illegal length")
	}
	if utils.IsContainIllegalChar(rq.Name, []string{",", "~", "#", "$", "%"}) {
		v.SetError("name", "robot name contains illegal characters")
	}
}

// RobotRep ...
type RobotRep struct {
	Name  string `json:"name"`
	Token string `json:"token"`
}
