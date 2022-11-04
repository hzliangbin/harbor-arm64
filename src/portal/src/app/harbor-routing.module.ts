// Copyright Project Harbor Authors
//
// Licensed under the Apache License, Version 2.0 (the 'License');
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//    http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an 'AS IS' BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.
import { NgModule } from '@angular/core';
import { RouterModule, Routes } from '@angular/router';

import { SystemAdminGuard } from './shared/route/system-admin-activate.service';
import { AuthCheckGuard } from './shared/route/auth-user-activate.service';
import { SignInGuard } from './shared/route/sign-in-guard-activate.service';
import { MemberGuard } from './shared/route/member-guard-activate.service';
import { MemberPermissionGuard } from './shared/route/member-permission-guard-activate.service';
import { OidcGuard } from './shared/route/oidc-guard-active.service';

import { PageNotFoundComponent } from './shared/not-found/not-found.component';
import { HarborShellComponent } from './base/harbor-shell/harbor-shell.component';
import { ConfigurationComponent } from './config/config.component';
import { DevCenterComponent } from './dev-center/dev-center.component';
import { GcPageComponent } from './gc-page/gc-page.component';
import { VulnerabilityPageComponent } from './vulnerability-page/vulnerability-page.component';

import { UserComponent } from './user/user.component';
import { SignInComponent } from './sign-in/sign-in.component';
import { ResetPasswordComponent } from './account/password-setting/reset-password/reset-password.component';
import { GroupComponent } from './group/group.component';

import { TotalReplicationPageComponent } from './replication/total-replication/total-replication-page.component';
import { ReplicationTasksPageComponent } from './replication/replication-tasks-page/replication-tasks-page.component';

import { DestinationPageComponent } from './replication/destination/destination-page.component';

import { AuditLogComponent } from './log/audit-log.component';
import { LogPageComponent } from './log/log-page.component';

import { RepositoryPageComponent } from './repository/repository-page.component';
import { TagRepositoryComponent } from './repository/tag-repository/tag-repository.component';
import { TagDetailPageComponent } from './repository/tag-detail/tag-detail-page.component';
import { LeavingRepositoryRouteDeactivate } from './shared/route/leaving-repository-deactivate.service';

import { ProjectComponent } from './project/project.component';
import { ProjectDetailComponent } from './project/project-detail/project-detail.component';
import { MemberComponent } from './project/member/member.component';
import { RobotAccountComponent } from './project/robot-account/robot-account.component';
import { WebhookComponent } from './project/webhook/webhook.component';
import { ProjectLabelComponent } from './project/project-label/project-label.component';
import { ProjectConfigComponent } from './project/project-config/project-config.component';
import { ProjectRoutingResolver } from './project/project-routing-resolver.service';
import { ListChartsComponent } from './project/helm-chart/list-charts.component';
import { ListChartVersionsComponent } from './project/helm-chart/list-chart-versions/list-chart-versions.component';
import { HelmChartDetailComponent } from './project/helm-chart/helm-chart-detail/chart-detail.component';
import { OidcOnboardComponent } from './oidc-onboard/oidc-onboard.component';
import { LicenseComponent } from './license/license.component';
import { SummaryComponent } from './project/summary/summary.component';
import { TagRetentionComponent } from './project/tag-retention/tag-retention.component';
import { ImmutableTagComponent } from './project/immutable-tag/immutable-tag.component';
import { USERSTATICPERMISSION, VulnerabilityConfigComponent } from '@harbor/ui';
import { ScannerComponent } from "./project/scanner/scanner.component";
import { InterrogationServicesComponent } from "./interrogation-services/interrogation-services.component";
import { ConfigurationScannerComponent } from "./config/scanner/config-scanner.component";
import { LabelsComponent } from "./labels/labels.component";
import { ProjectQuotasComponent } from "./project-quotas/project-quotas.component";


const harborRoutes: Routes = [
  { path: '', redirectTo: 'harbor', pathMatch: 'full' },
  { path: 'reset_password', component: ResetPasswordComponent },
  {
    path: 'devcenter',
    component: DevCenterComponent
  },
  {
    path: 'oidc-onboard',
    component: OidcOnboardComponent,
    canActivate: [OidcGuard, SignInGuard]
  },
  {
    path: 'license',
    component: LicenseComponent
  },
  {
    path: 'harbor/sign-in',
    component: SignInComponent,
    canActivate: [SignInGuard]
  },
  {
    path: 'harbor',
    component: HarborShellComponent,
    // canActivate: [AuthCheckGuard],
    canActivateChild: [AuthCheckGuard],
    children: [
      { path: '', redirectTo: 'projects', pathMatch: 'full' },
      {
        path: 'projects',
        component: ProjectComponent
      },
      {
        path: 'logs',
        component: LogPageComponent
      },
      {
        path: 'users',
        component: UserComponent,
        canActivate: [SystemAdminGuard]
      },
      {
        path: 'groups',
        component: GroupComponent,
        canActivate: [SystemAdminGuard]
      },
      {
        path: 'registries',
        component: DestinationPageComponent,
        canActivate: [SystemAdminGuard]
      },
      {
        path: 'replications',
        component: TotalReplicationPageComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard]
      },
      {
        path: 'interrogation-services',
        component: InterrogationServicesComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard],
        children: [
          {
            path: 'scanners',
            component: ConfigurationScannerComponent
          },
          {
            path: 'vulnerability',
            component: VulnerabilityConfigComponent
          },
          {
            path: '',
            redirectTo: 'scanners',
            pathMatch: 'full'
          },
        ]
      },
      {
        path: 'labels',
        component: LabelsComponent,
        canActivate: [SystemAdminGuard],
      },
      {
        path: 'project-quotas',
        component: ProjectQuotasComponent,
        canActivate: [SystemAdminGuard],
      },
      {
        path: 'replications/:id/:tasks',
        component: ReplicationTasksPageComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard]
      },
      {
        path: 'tags/:id/:repo',
        component: TagRepositoryComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        }
      },
      {
        path: 'projects/:id/repositories/:repo',
        component: TagRepositoryComponent,
        canActivate: [MemberGuard],
        canDeactivate: [LeavingRepositoryRouteDeactivate],
        resolve: {
          projectResolver: ProjectRoutingResolver
        }
      },
      {
        path: 'projects/:id/repositories/:repo/tags/:tag',
        component: TagDetailPageComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        }
      },
      {
        path: 'projects/:id/helm-charts/:chart/versions',
        component: ListChartVersionsComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        }
      },
      {
        path: 'projects/:id/helm-charts/:chart/versions/:version',
        component: HelmChartDetailComponent,
        canActivate: [MemberGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        }
      },
      {
        path: 'projects/:id',
        component: ProjectDetailComponent,
        canActivate: [MemberGuard],
        canActivateChild: [MemberPermissionGuard],
        resolve: {
          projectResolver: ProjectRoutingResolver
        },
        children: [
          {
            path: 'summary',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.PROJECT.KEY,
                action: USERSTATICPERMISSION.PROJECT.VALUE.READ
              }
            },
            component: SummaryComponent
          },
          {
            path: 'repositories',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.REPOSITORY.KEY,
                action: USERSTATICPERMISSION.REPOSITORY.VALUE.LIST
              }
            },
            component: RepositoryPageComponent,
          },
          {
            path: 'helm-charts',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.HELM_CHART.KEY,
                action: USERSTATICPERMISSION.HELM_CHART.VALUE.LIST
              }
            },
            component: ListChartsComponent
          },
          {
            path: 'repositories/:repo/tags',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.REPOSITORY.KEY,
                action: USERSTATICPERMISSION.REPOSITORY.VALUE.LIST
              }
            },
            component: TagRepositoryComponent
          },
          {
            path: 'members',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.MEMBER.KEY,
                action: USERSTATICPERMISSION.MEMBER.VALUE.LIST
              }
            },
            component: MemberComponent
          },
          {
            path: 'logs',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.LOG.KEY,
                action: USERSTATICPERMISSION.LOG.VALUE.LIST
              }
            },
            component: AuditLogComponent
          },
          {
            path: 'labels',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.LABEL.KEY,
                action: USERSTATICPERMISSION.LABEL.VALUE.CREATE
              }
            },
            component: ProjectLabelComponent
          },
          {
            path: 'configs',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.CONFIGURATION.KEY,
                action: USERSTATICPERMISSION.CONFIGURATION.VALUE.READ
              }
            },
            component: ProjectConfigComponent
          },
          {
            path: 'robot-account',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.ROBOT.KEY,
                action: USERSTATICPERMISSION.ROBOT.VALUE.LIST
              }
            },
            component: RobotAccountComponent
          },
          {
            path: 'tag-retention',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.TAG_RETENTION.KEY,
                action: USERSTATICPERMISSION.TAG_RETENTION.VALUE.READ
              }
            },
            component: TagRetentionComponent
          },
          {
            path: 'immutable-tag',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.TAG_RETENTION.KEY,
                action: USERSTATICPERMISSION.TAG_RETENTION.VALUE.READ
              }
            },
            component: ImmutableTagComponent
          },
          {
            path: 'webhook',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.WEBHOOK.KEY,
                action: USERSTATICPERMISSION.WEBHOOK.VALUE.LIST
              }
            },
            component: WebhookComponent
          },
          {
            path: 'scanner',
            data: {
              permissionParam: {
                resource: USERSTATICPERMISSION.SCANNER.KEY,
                action: USERSTATICPERMISSION.SCANNER.VALUE.READ
              }
            },
            component: ScannerComponent
          }
        ]
      },
      {
        path: 'configs',
        component: ConfigurationComponent,
        canActivate: [SystemAdminGuard]
      },
      {
        path: 'gc',
        component: GcPageComponent,
        canActivate: [SystemAdminGuard]
      },
      {
        path: 'registry',
        component: DestinationPageComponent,
        canActivate: [SystemAdminGuard],
        canActivateChild: [SystemAdminGuard]
      }
    ]
  },
  { path: '**', component: PageNotFoundComponent }
];

@NgModule({
  imports: [
    RouterModule.forRoot(harborRoutes, { onSameUrlNavigation: 'reload' })
  ],
  exports: [RouterModule]
})
export class HarborRoutingModule {}
