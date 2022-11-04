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

package scan

import "time"

// Report of the scan.
// Identified by the `digest`, `registration_uuid` and `mime_type`.
type Report struct {
	ID               int64     `orm:"pk;auto;column(id)"`
	UUID             string    `orm:"unique;column(uuid)"`
	Digest           string    `orm:"column(digest)"`
	RegistrationUUID string    `orm:"column(registration_uuid)"`
	MimeType         string    `orm:"column(mime_type)"`
	JobID            string    `orm:"column(job_id)"`
	TrackID          string    `orm:"column(track_id)"`
	Requester        string    `orm:"column(requester)"`
	Status           string    `orm:"column(status)"`
	StatusCode       int       `orm:"column(status_code)"`
	StatusRevision   int64     `orm:"column(status_rev)"`
	Report           string    `orm:"column(report);type(json)"`
	StartTime        time.Time `orm:"column(start_time);auto_now_add;type(datetime)"`
	EndTime          time.Time `orm:"column(end_time);type(datetime)"`
}

// TableName for Report
func (r *Report) TableName() string {
	return "scan_report"
}

// TableUnique for Report
func (r *Report) TableUnique() [][]string {
	return [][]string{
		{"uuid"},
		{"digest", "registration_uuid", "mime_type"},
	}
}
