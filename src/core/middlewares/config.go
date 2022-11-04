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

package middlewares

// const variables
const (
	CHART            = "chart"
	READONLY         = "readonly"
	URL              = "url"
	MUITIPLEMANIFEST = "manifest"
	LISTREPO         = "listrepo"
	CONTENTTRUST     = "contenttrust"
	VULNERABLE       = "vulnerable"
	SIZEQUOTA        = "sizequota"
	COUNTQUOTA       = "countquota"
	IMMUTABLE        = "immutable"
	REGTOKEN         = "regtoken"
)

// ChartMiddlewares middlewares for chart server
var ChartMiddlewares = []string{CHART}

// Middlewares with sequential organization
var Middlewares = []string{READONLY, URL, REGTOKEN, MUITIPLEMANIFEST, LISTREPO, CONTENTTRUST, VULNERABLE, SIZEQUOTA, IMMUTABLE, COUNTQUOTA}

// MiddlewaresLocal ...
var MiddlewaresLocal = []string{SIZEQUOTA, IMMUTABLE, COUNTQUOTA}
