# Copyright Project Harbor Authors
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#	http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License

*** Settings ***
Documentation  This resource wrap test case body

*** Variables ***

*** Keywords ***
Body Of Manage project publicity
    Init Chrome Driver
    ${d}=    Get Current Date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Create An New Project  project${d}  public=true

    Push image  ${ip}  user007  Test1@34  project${d}  hello-world:latest
    Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Search Private Projects
    Project Should Not Display  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Private  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Not Display  project${d}
    Cannot Pull image  ${ip}  user008  Test1@34  project${d}  hello-world:latest

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user007  Test1@34
    Make Project Public  project${d}

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user008  Test1@34
    Project Should Display  project${d}
    Close Browser

Body Of Scan A Tag In The Repo
    Init Chrome Driver
    ${d}=  get current date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user023  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Push Image  ${ip}  user023  Test1@34  project${d}  hello-world
    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Scan Repo  latest  Succeed
    Summary Chart Should Display  latest
    Pull Image  ${ip}  user023  Test1@34  project${d}  hello-world
    # Edit Repo Info
    Close Browser

Body Of List Helm Charts
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user027  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}

    Switch To Project Charts
    Upload Chart files
    Go Into Chart Version  ${prometheus_chart_name}
    Retry Wait Until Page Contains  ${prometheus_chart_version}
    Go Into Chart Detail  ${prometheus_chart_version}

    # Summary tab
    Retry Wait Until Page Contains Element  ${summary_markdown}
    Retry Wait Until Page Contains Element  ${summary_container}

    # Dependency tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_dependency}  Retry Wait Until Page Contains Element  ${dependency_content}

    # Values tab
    Retry Double Keywords When Error  Retry Element Click  xpath=${detail_value}  Retry Wait Until Page Contains Element  ${value_content}

    Go Into Project  project${d}  has_image=${false}
    Switch To Project Charts
    Multi-delete Chart Files  ${prometheus_chart_name}  ${harbor_chart_name}
    Close Browser

Body Of Admin Push Signed Image
    [Arguments]  ${image}=tomcat  ${with_remove}=${false}
    Enable Notary Client

    Docker Pull  ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}
    ${rc}  ${output}=  Run And Return Rc And Output  ./tests/robot-cases/Group0-Util/notary-push-image.sh ${ip} library ${image} latest ${notaryServerEndpoint} ${LOCAL_REGISTRY}/${LOCAL_REGISTRY_NAMESPACE}/${image}:latest
    Log  ${output}
    Should Be Equal As Integers  ${rc}  0

    ${rc}  ${output}=  Run And Return Rc And Output  curl -u admin:Harbor12345 -s --insecure -H "Content-Type: application/json" -X GET "https://${ip}/api/repositories/library/${image}/signatures"
    Log To Console  ${output}
    Should Be Equal As Integers  ${rc}  0
    Should Contain  ${output}  sha256

    Run Keyword If  ${with_remove} == ${true}  Remove Notary Signature  ${ip}  ${image}

Delete A Project Without Sign In Harbor
    [Arguments]  ${harbor_ip}=${ip}  ${username}=${HARBOR_ADMIN}  ${password}=${HARBOR_PASSWORD}
    ${d}=    Get Current Date    result_format=%m%s
    Create An New Project  project${d}
    Push Image  ${harbor_ip}  ${username}  ${password}  project${d}  hello-world
    Project Should Not Be Deleted  project${d}
    Go Into Project  project${d}
    Delete Repo  project${d}
    Navigate To Projects
    Project Should Be Deleted  project${d}

Manage Project Member Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}  ${test_user1}=user005  ${test_user2}=user006  ${is_oidc_mode}=${false}
    ${d}=    Get current Date  result_format=%m%s
    Create An New Project  project${d}
    Push image  ip=${ip}  user=${sign_in_user}  pwd=${sign_in_pwd}  project=project${d}  image=hello-world
    Logout Harbor

    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Add  is_oidc_mode=${is_oidc_mode}
    User Should Be Guest  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Developer  is_oidc_mode=${is_oidc_mode}
    User Should Be Developer  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Admin  is_oidc_mode=${is_oidc_mode}
    User Should Be Admin  ${test_user1}  ${sign_in_pwd}  project${d}  ${test_user2}  is_oidc_mode=${is_oidc_mode}
    Change User Role In Project  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Master  is_oidc_mode=${is_oidc_mode}
    User Should Be Master  ${test_user1}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}
    Manage Project Member  ${sign_in_user}  ${sign_in_pwd}  project${d}  ${test_user1}  Remove  is_oidc_mode=${is_oidc_mode}
    User Should Not Be A Member Of Project  ${test_user1}  ${sign_in_pwd}  project${d}    is_oidc_mode=${is_oidc_mode}
    Push image  ip=${ip}  user=${sign_in_user}  pwd=${sign_in_pwd}  project=project${d}  image=hello-world
    User Should Be Guest  ${test_user2}  ${sign_in_pwd}  project${d}  is_oidc_mode=${is_oidc_mode}

Helm CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project  project${d}
    Helm Repo Add  ${HARBOR_URL}  ${sign_in_user}  ${sign_in_pwd}  project_name=project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}
    Go Into Project  project${d}  has_image=${false}
    Switch To Project Charts
    Retry Double Keywords When Error  Go Into Chart Version  ${harbor_chart_name}  Retry Wait Until Page Contains  ${harbor_chart_version}
    Capture Page Screenshot

Helm3 CLI Push Without Sign In Harbor
    [Arguments]  ${sign_in_user}  ${sign_in_pwd}
    ${d}=   Get Current Date    result_format=%m%s
    Create An New Project  project${d}
    Helm Repo Push  ${sign_in_user}  ${sign_in_pwd}  ${harbor_chart_filename}  helm_repo_name=${HARBOR_URL}/chartrepo/project${d}  helm_cmd=helm3
    Go Into Project  project${d}  has_image=${false}
    Switch To Project Charts
    Retry Double Keywords When Error  Go Into Chart Version  ${harbor_chart_name}  Retry Wait Until Page Contains  ${harbor_chart_version}
    Capture Page Screenshot

Body Of Replication Of Push Images to Registry Triggered By Event
    [Arguments]  ${provider}  ${endpoint}  ${username}  ${pwd}  ${dest_namespace}
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    ${sha256}=  Set Variable  0e67625224c1da47cb3270e7a861a83e332f708d3d89dde0cbed432c94824d9a
    ${image}=  Set Variable  test_push_repli
    ${tag1}=  Set Variable  v1.1.0
    @{tags}   Create List  ${tag1}
    #login source
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project    project${d}
    Switch To Registries
    Create A New Endpoint    ${provider}    e${d}    ${endpoint}    ${username}    ${pwd}    Y
    Switch To Replication Manage
    Create A Rule With Existing Endpoint    rule${d}    push    project${d}/*    image    e${d}    ${dest_namespace}  mode=Event Based  del_remote=${true}
    Push Special Image To Project  project${d}  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  ${image}  tags=@{tags}  size=12
    Filter Replication Rule  rule${d}
    Select Rule  rule${d}
    Run Keyword If  '${provider}'=='docker-hub'  Docker Image Can Be Pulled  ${dest_namespace}/${image}:${tag1}   times=3
    Executions Result Count Should Be  Succeed  event_based  1
    Go Into Project  project${d}
    Delete Repo  project${d}
    Run Keyword If  '${provider}'=='docker-hub'  Docker Image Can Not Be Pulled  ${dest_namespace}/${image}:${tag1}
    Switch To Replication Manage
    Filter Replication Rule  rule${d}
    Select Rule  rule${d}
    Executions Result Count Should Be  Succeed  event_based  2