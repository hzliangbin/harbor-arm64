// Copyright 2018 Project Harbor Authors
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
	"regexp"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/goharbor/harbor/src/common"
	"github.com/goharbor/harbor/src/common/dao"
	"github.com/goharbor/harbor/src/common/dao/project"
	"github.com/goharbor/harbor/src/common/models"
	"github.com/goharbor/harbor/src/common/quota"
	"github.com/goharbor/harbor/src/common/rbac"
	"github.com/goharbor/harbor/src/common/utils"
	errutil "github.com/goharbor/harbor/src/common/utils/error"
	"github.com/goharbor/harbor/src/common/utils/log"
	"github.com/goharbor/harbor/src/core/config"
	"github.com/goharbor/harbor/src/pkg/scan/vuln"
	"github.com/goharbor/harbor/src/pkg/types"
	"github.com/pkg/errors"
)

type deletableResp struct {
	Deletable bool   `json:"deletable"`
	Message   string `json:"message"`
}

// ProjectAPI handles request to /api/projects/{} /api/projects/{}/logs
type ProjectAPI struct {
	BaseController
	project *models.Project
}

const projectNameMaxLen int = 255
const projectNameMinLen int = 1
const restrictedNameChars = `[a-z0-9]+(?:[._-][a-z0-9]+)*`

// Prepare validates the URL and the user
func (p *ProjectAPI) Prepare() {
	p.BaseController.Prepare()
	if len(p.GetStringFromPath(":id")) != 0 {
		id, err := p.GetInt64FromPath(":id")
		if err != nil || id <= 0 {
			text := "invalid project ID: "
			if err != nil {
				text += err.Error()
			} else {
				text += fmt.Sprintf("%d", id)
			}
			p.SendBadRequestError(errors.New(text))
			return
		}

		project, err := p.ProjectMgr.Get(id)
		if err != nil {
			p.ParseAndHandleError(fmt.Sprintf("failed to get project %d", id), err)
			return
		}

		if project == nil {
			p.handleProjectNotFound(id)
			return
		}

		p.project = project
	}
}

func (p *ProjectAPI) requireAccess(action rbac.Action, subresource ...rbac.Resource) bool {
	if len(subresource) == 0 {
		subresource = append(subresource, rbac.ResourceSelf)
	}

	return p.RequireProjectAccess(p.project.ProjectID, action, subresource...)
}

// Post ...
func (p *ProjectAPI) Post() {
	if !p.SecurityCtx.IsAuthenticated() {
		p.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}
	var onlyAdmin bool
	var err error
	if config.WithAdmiral() {
		onlyAdmin = true
	} else {
		onlyAdmin, err = config.OnlyAdminCreateProject()
		if err != nil {
			log.Errorf("failed to determine whether only admin can create projects: %v", err)
			p.SendInternalServerError(fmt.Errorf("failed to determine whether only admin can create projects: %v", err))
			return
		}
	}

	if onlyAdmin && !(p.SecurityCtx.IsSysAdmin() || p.SecurityCtx.IsSolutionUser()) {
		log.Errorf("Only sys admin can create project")
		p.SendForbiddenError(errors.New("Only system admin can create project"))
		return
	}
	var pro *models.ProjectRequest
	if err := p.DecodeJSONReq(&pro); err != nil {
		p.SendBadRequestError(err)
		return
	}

	err = validateProjectReq(pro)
	if err != nil {
		log.Errorf("Invalid project request, error: %v", err)
		p.SendBadRequestError(fmt.Errorf("invalid request: %v", err))
		return
	}

	var hardLimits types.ResourceList
	if config.QuotaPerProjectEnable() {
		setting, err := config.QuotaSetting()
		if err != nil {
			log.Errorf("failed to get quota setting: %v", err)
			p.SendInternalServerError(fmt.Errorf("failed to get quota setting: %v", err))
			return
		}

		if !p.SecurityCtx.IsSysAdmin() {
			pro.CountLimit = &setting.CountPerProject
			pro.StorageLimit = &setting.StoragePerProject
		}

		hardLimits, err = projectQuotaHardLimits(pro, setting)
		if err != nil {
			log.Errorf("Invalid project request, error: %v", err)
			p.SendBadRequestError(fmt.Errorf("invalid request: %v", err))
			return
		}
	}

	exist, err := p.ProjectMgr.Exists(pro.Name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to check the existence of project %s",
			pro.Name), err)
		return
	}
	if exist {
		p.SendConflictError(errors.New("conflict project"))
		return
	}

	if pro.Metadata == nil {
		pro.Metadata = map[string]string{}
	}
	// accept the "public" property to make replication work well with old versions(<=1.2.0)
	if pro.Public != nil && len(pro.Metadata[models.ProMetaPublic]) == 0 {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(*pro.Public == 1)
	}

	// populate public metadata as false if it isn't set
	if _, ok := pro.Metadata[models.ProMetaPublic]; !ok {
		pro.Metadata[models.ProMetaPublic] = strconv.FormatBool(false)
	}
	// populate

	owner := p.SecurityCtx.GetUsername()
	// set the owner as the system admin when the API being called by replication
	// it's a solution to workaround the restriction of project creation API:
	// only normal users can create projects
	if p.SecurityCtx.IsSolutionUser() {
		user, err := dao.GetUser(models.User{
			UserID: 1,
		})
		if err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to get the user 1: %v", err))
			return
		}
		owner = user.Username
	}
	projectID, err := p.ProjectMgr.Create(&models.Project{
		Name:      pro.Name,
		OwnerName: owner,
		Metadata:  pro.Metadata,
	})
	if err != nil {
		if err == errutil.ErrDupProject {
			log.Debugf("conflict %s", pro.Name)
			p.SendConflictError(fmt.Errorf("conflict %s", pro.Name))
		} else {
			p.ParseAndHandleError("failed to add project", err)
		}
		return
	}

	if config.QuotaPerProjectEnable() {
		quotaMgr, err := quota.NewManager("project", strconv.FormatInt(projectID, 10))
		if err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to get quota manager: %v", err))
			return
		}
		if _, err := quotaMgr.NewQuota(hardLimits); err != nil {
			p.SendInternalServerError(fmt.Errorf("failed to create quota for project: %v", err))
			return
		}
	}

	go func() {
		if err = dao.AddAccessLog(
			models.AccessLog{
				Username:  p.SecurityCtx.GetUsername(),
				ProjectID: projectID,
				RepoName:  pro.Name + "/",
				RepoTag:   "N/A",
				Operation: "create",
				OpTime:    time.Now(),
			}); err != nil {
			log.Errorf("failed to add access log: %v", err)
		}
	}()

	p.Redirect(http.StatusCreated, strconv.FormatInt(projectID, 10))
}

// Head ...
func (p *ProjectAPI) Head() {

	if !p.SecurityCtx.IsAuthenticated() {
		p.SendUnAuthorizedError(errors.New("Unauthorized"))
		return
	}

	name := p.GetString("project_name")
	if len(name) == 0 {
		p.SendBadRequestError(errors.New("project_name is needed"))
		return
	}

	project, err := p.ProjectMgr.Get(name)
	if err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to get project %s", name), err)
		return
	}

	if project == nil {
		p.SendNotFoundError(fmt.Errorf("project %s not found", name))
		return
	}
}

// Get ...
func (p *ProjectAPI) Get() {
	if !p.requireAccess(rbac.ActionRead) {
		return
	}

	err := p.populateProperties(p.project)
	if err != nil {
		log.Errorf("populate project properties failed with : %+v", err)
	}

	p.Data["json"] = p.project
	p.ServeJSON()
}

// Delete ...
func (p *ProjectAPI) Delete() {
	if !p.requireAccess(rbac.ActionDelete) {
		return
	}

	result, err := p.deletable(p.project.ProjectID)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}
	if !result.Deletable {
		p.SendPreconditionFailedError(errors.New(result.Message))
		return
	}

	quotaMgr, err := quota.NewManager("project", strconv.FormatInt(p.project.ProjectID, 10))
	if err != nil {
		p.SendInternalServerError(fmt.Errorf("failed to get quota manager: %v", err))
		return
	}
	if err := quotaMgr.DeleteQuota(); err != nil {
		p.SendInternalServerError(fmt.Errorf("failed to delete quota for project: %v", err))
		return
	}

	if err = p.ProjectMgr.Delete(p.project.ProjectID); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to delete project %d", p.project.ProjectID), err)
		return
	}

	go func() {
		if err := dao.AddAccessLog(models.AccessLog{
			Username:  p.SecurityCtx.GetUsername(),
			ProjectID: p.project.ProjectID,
			RepoName:  p.project.Name + "/",
			RepoTag:   "N/A",
			Operation: "delete",
			OpTime:    time.Now(),
		}); err != nil {
			log.Errorf("failed to add access log: %v", err)
		}
	}()
}

// Deletable ...
func (p *ProjectAPI) Deletable() {
	if !p.requireAccess(rbac.ActionDelete) {
		return
	}

	result, err := p.deletable(p.project.ProjectID)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to check the deletable of project %d: %v", p.project.ProjectID, err))
		return
	}

	p.Data["json"] = result
	p.ServeJSON()
}

func (p *ProjectAPI) deletable(projectID int64) (*deletableResp, error) {
	count, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{projectID},
	})
	if err != nil {
		return nil, err
	}

	if count > 0 {
		return &deletableResp{
			Deletable: false,
			Message:   "the project contains repositories, can not be deleted",
		}, nil
	}

	// Check helm charts number
	if config.WithChartMuseum() {
		charts, err := chartController.ListCharts(p.project.Name)
		if err != nil {
			return nil, err
		}

		if len(charts) > 0 {
			return &deletableResp{
				Deletable: false,
				Message:   "the project contains helm charts, can not be deleted",
			}, nil
		}
	}

	return &deletableResp{
		Deletable: true,
	}, nil
}

// List ...
func (p *ProjectAPI) List() {
	// query strings
	page, size, err := p.GetPaginationParams()
	if err != nil {
		p.SendBadRequestError(err)
		return
	}
	query := &models.ProjectQueryParam{
		Name:  p.GetString("name"),
		Owner: p.GetString("owner"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	public := p.GetString("public")
	if len(public) > 0 {
		pub, err := strconv.ParseBool(public)
		if err != nil {
			p.SendBadRequestError(fmt.Errorf("invalid public: %s", public))
			return
		}
		query.Public = &pub
	}

	// standalone, filter projects according to the privilleges of the user first
	if !config.WithAdmiral() {
		var projects []*models.Project
		if !p.SecurityCtx.IsAuthenticated() {
			// not login, only get public projects
			pros, err := p.ProjectMgr.GetPublic()
			if err != nil {
				p.SendInternalServerError(fmt.Errorf("failed to get public projects: %v", err))
				return
			}
			projects = []*models.Project{}
			projects = append(projects, pros...)
		} else {
			if !(p.SecurityCtx.IsSysAdmin() || p.SecurityCtx.IsSolutionUser()) {
				projects = []*models.Project{}
				// login, but not system admin or solution user, get public projects and
				// projects that the user is member of
				pros, err := p.ProjectMgr.GetPublic()
				if err != nil {
					p.SendInternalServerError(fmt.Errorf("failed to get public projects: %v", err))
					return
				}
				projects = append(projects, pros...)
				mps, err := p.SecurityCtx.GetMyProjects()
				if err != nil {
					p.SendInternalServerError(fmt.Errorf("failed to list projects: %v", err))
					return
				}
				projects = append(projects, mps...)
			}
		}
		// Query projects by user group

		if projects != nil {
			projectIDs := []int64{}
			for _, project := range projects {
				projectIDs = append(projectIDs, project.ProjectID)
			}
			query.ProjectIDs = projectIDs
		}
	}

	result, err := p.ProjectMgr.List(query)
	if err != nil {
		p.ParseAndHandleError("failed to list projects", err)
		return
	}

	for _, project := range result.Projects {
		err = p.populateProperties(project)
		if err != nil {
			log.Errorf("populate project properties failed %v", err)
		}
	}
	p.SetPaginationHeader(result.Total, page, size)
	p.Data["json"] = result.Projects
	p.ServeJSON()
}

func (p *ProjectAPI) populateProperties(project *models.Project) error {
	// Transform the severity to severity of CVSS v3.0 Ratings
	if severity, ok := project.GetMetadata(models.ProMetaSeverity); ok {
		project.SetMetadata(models.ProMetaSeverity, strings.ToLower(vuln.ParseSeverityVersion3(severity).String()))
	}

	if p.SecurityCtx.IsAuthenticated() {
		roles := p.SecurityCtx.GetProjectRoles(project.ProjectID)
		project.RoleList = roles
		project.Role = highestRole(roles)
	}

	total, err := dao.GetTotalOfRepositories(&models.RepositoryQuery{
		ProjectIDs: []int64{project.ProjectID},
	})
	if err != nil {
		err = errors.Wrap(err, fmt.Sprintf("get repo count of project %d failed", project.ProjectID))
		return err
	}

	project.RepoCount = total

	// Populate chart count property
	if config.WithChartMuseum() {
		count, err := chartController.GetCountOfCharts([]string{project.Name})
		if err != nil {
			err = errors.Wrap(err, fmt.Sprintf("get chart count of project %d failed", project.ProjectID))
			return err
		}

		project.ChartCount = count
	}
	return nil
}

// Put ...
func (p *ProjectAPI) Put() {
	if !p.requireAccess(rbac.ActionUpdate) {
		return
	}

	var req *models.ProjectRequest
	if err := p.DecodeJSONReq(&req); err != nil {
		p.SendBadRequestError(err)
		return
	}

	if err := p.ProjectMgr.Update(p.project.ProjectID,
		&models.Project{
			Metadata:     req.Metadata,
			CVEWhitelist: req.CVEWhitelist,
		}); err != nil {
		p.ParseAndHandleError(fmt.Sprintf("failed to update project %d",
			p.project.ProjectID), err)
		return
	}
}

// Logs ...
func (p *ProjectAPI) Logs() {
	if !p.requireAccess(rbac.ActionList, rbac.ResourceLog) {
		return
	}

	page, size, err := p.GetPaginationParams()
	if err != nil {
		p.SendBadRequestError(err)
		return
	}
	query := &models.LogQueryParam{
		ProjectIDs: []int64{p.project.ProjectID},
		Username:   p.GetString("username"),
		Repository: p.GetString("repository"),
		Tag:        p.GetString("tag"),
		Operations: p.GetStrings("operation"),
		Pagination: &models.Pagination{
			Page: page,
			Size: size,
		},
	}

	timestamp := p.GetString("begin_timestamp")
	if len(timestamp) > 0 {
		t, err := utils.ParseTimeStamp(timestamp)
		if err != nil {
			p.SendBadRequestError(fmt.Errorf("invalid begin_timestamp: %s", timestamp))
			return
		}
		query.BeginTime = t
	}

	timestamp = p.GetString("end_timestamp")
	if len(timestamp) > 0 {
		t, err := utils.ParseTimeStamp(timestamp)
		if err != nil {
			p.SendBadRequestError(fmt.Errorf("invalid end_timestamp: %s", timestamp))
			return
		}
		query.EndTime = t
	}

	total, err := dao.GetTotalOfAccessLogs(query)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to get total of access log: %v", err))
		return
	}

	logs, err := dao.GetAccessLogs(query)
	if err != nil {
		p.SendInternalServerError(fmt.Errorf(
			"failed to get access log: %v", err))
		return
	}

	p.SetPaginationHeader(total, page, size)
	p.Data["json"] = logs
	p.ServeJSON()
}

// Summary returns the summary of the project
func (p *ProjectAPI) Summary() {
	if !p.requireAccess(rbac.ActionRead) {
		return
	}

	if err := p.populateProperties(p.project); err != nil {
		log.Warningf("populate project properties failed with : %+v", err)
	}

	summary := &models.ProjectSummary{
		RepoCount:  p.project.RepoCount,
		ChartCount: p.project.ChartCount,
	}

	var fetchSummaries []func(int64, *models.ProjectSummary)

	if hasPerm, _ := p.HasProjectPermission(p.project.ProjectID, rbac.ActionRead, rbac.ResourceQuota); hasPerm {
		fetchSummaries = append(fetchSummaries, getProjectQuotaSummary)
	}

	if hasPerm, _ := p.HasProjectPermission(p.project.ProjectID, rbac.ActionList, rbac.ResourceMember); hasPerm {
		fetchSummaries = append(fetchSummaries, getProjectMemberSummary)
	}

	var wg sync.WaitGroup
	for _, fn := range fetchSummaries {
		fn := fn

		wg.Add(1)
		go func() {
			defer wg.Done()
			fn(p.project.ProjectID, summary)
		}()
	}
	wg.Wait()

	p.Data["json"] = summary
	p.ServeJSON()
}

// TODO move this to pa ckage models
func validateProjectReq(req *models.ProjectRequest) error {
	pn := req.Name
	if utils.IsIllegalLength(pn, projectNameMinLen, projectNameMaxLen) {
		return fmt.Errorf("Project name %s is illegal in length. (greater than %d or less than %d)", pn, projectNameMaxLen, projectNameMinLen)
	}
	validProjectName := regexp.MustCompile(`^` + restrictedNameChars + `$`)
	legal := validProjectName.MatchString(pn)
	if !legal {
		return fmt.Errorf("project name is not in lower case or contains illegal characters")
	}

	metas, err := validateProjectMetadata(req.Metadata)
	if err != nil {
		return err
	}

	req.Metadata = metas
	return nil
}

func projectQuotaHardLimits(req *models.ProjectRequest, setting *models.QuotaSetting) (types.ResourceList, error) {
	hardLimits := types.ResourceList{}
	if req.CountLimit != nil {
		hardLimits[types.ResourceCount] = *req.CountLimit
	} else {
		hardLimits[types.ResourceCount] = setting.CountPerProject
	}

	if req.StorageLimit != nil {
		hardLimits[types.ResourceStorage] = *req.StorageLimit
	} else {
		hardLimits[types.ResourceStorage] = setting.StoragePerProject
	}

	if err := quota.Validate("project", hardLimits); err != nil {
		return nil, err
	}

	return hardLimits, nil
}

func getProjectQuotaSummary(projectID int64, summary *models.ProjectSummary) {
	if !config.QuotaPerProjectEnable() {
		log.Debug("Quota per project disabled")
		return
	}

	quotas, err := dao.ListQuotas(&models.QuotaQuery{Reference: "project", ReferenceID: strconv.FormatInt(projectID, 10)})
	if err != nil {
		log.Debugf("failed to get quota for project: %d", projectID)
		return
	}

	if len(quotas) == 0 {
		log.Debugf("quota not found for project: %d", projectID)
		return
	}

	quota := quotas[0]

	summary.Quota.Hard, _ = types.NewResourceList(quota.Hard)
	summary.Quota.Used, _ = types.NewResourceList(quota.Used)
}

func getProjectMemberSummary(projectID int64, summary *models.ProjectSummary) {
	var wg sync.WaitGroup

	for _, e := range []struct {
		role  int
		count *int64
	}{
		{common.RoleProjectAdmin, &summary.ProjectAdminCount},
		{common.RoleMaster, &summary.MasterCount},
		{common.RoleDeveloper, &summary.DeveloperCount},
		{common.RoleGuest, &summary.GuestCount},
		{common.RoleLimitedGuest, &summary.LimitedGuestCount},
	} {
		wg.Add(1)
		go func(role int, count *int64) {
			defer wg.Done()

			total, err := project.GetTotalOfProjectMembers(projectID, role)
			if err != nil {
				log.Debugf("failed to get total of project members of role %d", role)
				return
			}

			*count = total
		}(e.role, e.count)
	}

	wg.Wait()
}

// Returns the highest role in the role list.
// This func should be removed once we deprecate the "current_user_role_id" in project API
// A user can have multiple roles and they may not have a strict ranking relationship
func highestRole(roles []int) int {
	if roles == nil {
		return 0
	}
	rolePower := map[int]int{
		common.RoleProjectAdmin: 50,
		common.RoleMaster:       40,
		common.RoleDeveloper:    30,
		common.RoleGuest:        20,
		common.RoleLimitedGuest: 10,
	}
	var highest, highestPower int
	for _, role := range roles {
		if p, ok := rolePower[role]; ok && p > highestPower {
			highest = role
			highestPower = p
		}
	}
	return highest
}
