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
Documentation  Harbor BATs
Resource  ../../resources/Util.robot
Default Tags  Nightly

*** Variables ***
${HARBOR_URL}  https://${ip}
${SSH_USER}  root
${HARBOR_ADMIN}  admin

*** Test Cases ***
Test Case - Sign With Admin
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Close Browser

Test Case - Create An New Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  test${d}
    Close Browser

Test Case - Delete A Project
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Delete A Project Without Sign In Harbor
    Close Browser

Test Case - Repo Size
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  library  alpine  2.6  2.6
    Go Into Project  library
    Go Into Repo  alpine
    Wait Until Page Contains  1.92MB
    Close Browser

Test Case - Staticsinfo
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${element}=  Set Variable  ${project_statistics_private_repository_icon}
    Wait Until Element Is Visible  ${element}
    ${privaterepocount1}=  Get Statics Private Repo
    ${privateprojcount1}=  Get Statics Private Project
    ${publicrepocount1}=  Get Statics Public Repo
    ${publicprojcount1}=  Get Statics Public Project
    ${totalrepocount1}=  Get Statics Total Repo
    ${totalprojcount1}=  Get Statics Total Project
    Create An New Project  private${d}
    Create An New Project  public${d}  true
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  private${d}  hello-world
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  public${d}  hello-world
    Reload Page
    ${privateprojcount}=  evaluate  ${privateprojcount1}+1
    ${privaterepocount}=  evaluate  ${privaterepocount1}+1
    ${publicprojcount}=  evaluate  ${publicprojcount1}+1
    ${publicrepocount}=  evaluate  ${publicrepocount1}+1
    ${totalrepocount}=  evaluate  ${totalrepocount1}+2
    ${totalprojcount}=  evaluate  ${totalprojcount1}+2
    Navigate To Projects
    Wait Until Element Is Visible  ${element}
    ${privaterepocountStr}=  Convert To String  ${privaterepocount}
    Wait Until Element Contains  ${element}  ${privaterepocountStr}
    ${privaterepocount2}=  Get Statics Private Repo
    ${privateprojcount2}=  get statics private project
    ${publicrepocount2}=  get statics public repo
    ${publicprojcount2}=  get statics public project
    ${totalrepocount2}=  get statics total repo
    ${totalprojcount2}=  get statics total project
    Should Be Equal As Integers  ${privateprojcount2}  ${privateprojcount}
    Should be equal as integers  ${privaterepocount2}  ${privaterepocount}
    Should be equal as integers  ${publicprojcount2}  ${publicprojcount}
    Should be equal as integers  ${publicrepocount2}  ${publicrepocount}
    Should be equal as integers  ${totalprojcount2}  ${totalprojcount}
    Should be equal as integers  ${totalrepocount2}  ${totalrepocount}

Test Case - Push Image
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  test${d}

    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  test${d}  hello-world:latest
    Go Into Project  test${d}
    Wait Until Page Contains  test${d}/hello-world

Test Case - Project Level Policy Public
    Init Chrome Driver
    ${d}=  Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Goto Project Config
    Click Project Public
    Save Project Config
    # Verify
    Public Should Be Selected
    # Project${d}  default should be private
    # Here logout and login to try avoid a bug only in autotest
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Retry Double Keywords When Error  Filter Project  project${d}  Project Should Be Public  project${d}
    Close Browser

Test Case - Verify Download Ca Link
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Settings
    Page Should Contain  Registry Root Certificate
    Close Browser

Test Case - Edit Email Settings
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}

    Switch To Email
    Config Email

    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}

    Switch To Email
    Verify Email

    Close Browser

Test Case - Edit Token Expire
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Settings
    Modify Token Expiration  20
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Settings
    Token Must Be Match  20

    #reset to default
    Modify Token Expiration  30
    Close Browser

Test Case - Create A New Labels
    Init Chrome Driver
    ${d}=    Get Current Date
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Labels
    Create New Labels  label_${d}
    Close Browser

Test Case - Update Label
   Init Chrome Driver
   ${d}=    Get Current Date

   Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
   Switch To System Labels
   Create New Labels  label_${d}
   Sleep  3
   ${d1}=    Get Current Date
   Update A Label  label_${d}
   Close Browser

Test Case - Delete Label
    Init Chrome Driver
    ${d}=    Get Current Date
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch To System Labels
    Create New Labels  label_${d}
    Sleep  3
    Delete A Label  label_${d}
    Close Browser

Test Case - User View Projects
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user001  Test1@34
    Create An New Project  test${d}1
    Create An New Project  test${d}2
    Create An New Project  test${d}3
    Switch To Log
    Wait Until Page Contains  test${d}1
    Wait Until Page Contains  test${d}2
    Wait Until Page Contains  test${d}3
    Close Browser

Test Case - User View Logs
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user002  Test1@34
    Create An New Project  project${d}

    Push image  ${ip}  user002  Test1@34  project${d}  busybox:latest
    Pull image  ${ip}  user002  Test1@34  project${d}  busybox:latest

    Go Into Project  project${d}
    Delete Repo  project${d}

    Sleep  3

    Go To Project Log
    Advanced Search Should Display

    Do Log Advanced Search
    Close Browser

Test Case - Manage Project Member
    Init Chrome Driver
    ${user}=    Set Variable    user004
    ${pwd}=    Set Variable    Test1@34
    Sign In Harbor  ${HARBOR_URL}  ${user}  ${pwd}
    Manage Project Member Without Sign In Harbor  ${user}  ${pwd}
    Close Browser

Test Case - Manage project publicity
    Body Of Manage project publicity

Test Case - Assign Sys Admin
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user009  Test1@34
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Switch to User Tag
    Assign User Admin  user009
    Logout Harbor
    Sign In Harbor  ${HARBOR_URL}  user009  Test1@34
    Administration Tag Should Display
    Close Browser

Test Case - Edit Project Creation
    # Create normal user and login
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user010  Test1@34
    Project Creation Should Display
    Logout Harbor

    Sleep  3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Set Pro Create Admin Only
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  user010  Test1@34
    Project Creation Should Not Display
    Logout Harbor

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Set Pro Create Every One
    Close browser

Test Case - Edit Repo Info
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user011  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user011  Test1@34  project${d}  hello-world
    Go Into Project  project${d}
    Go Into Repo  project${d}/hello-world
    Edit Repo Info
    Close Browser

Test Case - Delete Multi Project
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user012  Test1@34
    Create An New Project  projecta${d}
    Create An New Project  projectb${d}
    Push Image  ${ip}  user012  Test1@34  projecta${d}  hello-world
    Navigate To Projects
    Filter Object  project
    Retry Wait Element Not Visible  //clr-datagrid/div/div[2]
    @{project_list}  Create List  projecta  projectb
    Multi-delete Object  ${project_delete_btn}  @{project_list}
    # Verify delete project with image should not be deleted directly
    Delete Fail  projecta${d}
    Delete Success  projectb${d}
    Close Browser

Test Case - Delete Multi Repo
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user013  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user013  Test1@34  project${d}  hello-world
    Push Image  ${ip}  user013  Test1@34  project${d}  busybox
    Sleep  2
    Go Into Project  project${d}
    @{repo_list}  Create List  hello-world  busybox
    Multi-delete Object  ${repo_delete_btn}  @{repo_list}
    # Verify
    Delete Success  hello-world  busybox
    Close Browser

Test Case - Delete Multi Tag
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user014  Test1@34
    Create An New Project  project${d}
    Push Image With Tag  ${ip}  user014  Test1@34  project${d}  redis  3.2.10-alpine  3.2.10-alpine
    Push Image With Tag  ${ip}  user014  Test1@34  project${d}  redis  4.0.7-alpine  4.0.7-alpine
    Go Into Project  project${d}
    Go Into Repo  redis
    @{tag_list}  Create List  3.2.10-alpine  4.0.7-alpine
    Multi-delete object  ${tag_delete_btn}  @{tag_list}
    # Verify
    Delete Success  3.2.10-alpine  4.0.7-alpine
    Close Browser

Test Case - Delete Repo on CardView
    Init Chrome Driver
    ${d}=   Get Current Date  result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user015  Test1@34
    Create An New Project  project${d}
    Push Image  ${ip}  user015  Test1@34  project${d}  hello-world
    Push Image  ${ip}  user015  Test1@34  project${d}  busybox
    Go Into Project  project${d}
    Switch To CardView
    Delete Repo on CardView  busybox
    # Verify
    Delete Success  busybox
    Close Browser

Test Case - Delete Multi Member
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user016  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Switch To Member
    Add Guest Member To Project  user017
    Add Guest Member To Project  user018
    Multi-delete Member  user017  user018
    Delete Success  user017  user018
    Close Browser

Test Case - Project Admin Operate Labels
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user019  Test1@34
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Sleep  2
    # Add labels
    Switch To Project Label
    Create New Labels  label_${d}
    Sleep  2
    Update A Label  label_${d}
    Sleep  2
    Delete A Label  label_${d}
    Close Browser

Test Case - Project Admin Add Labels To Repo
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  user020  Test1@34
    Create An New Project  project${d}
    Push Image With Tag  ${ip}  user020  Test1@34  project${d}  redis  3.2.10-alpine  3.2.10-alpine
    Push Image With Tag  ${ip}  user020  Test1@34  project${d}  redis  4.0.7-alpine  4.0.7-alpine
    Go Into Project  project${d}
    Sleep  2
    # Add labels
    Switch To Project Label
    Create New Labels  label111
    Create New Labels  label22
    Sleep  2
    Switch To Project Repo
    Go Into Repo  project${d}/redis
    Add Labels To Tag  3.2.10-alpine  label111
    Add Labels To Tag  4.0.7-alpine  label22
    Filter Labels In Tags  label111  label22
    Close Browser

Test Case - Developer Operate Labels
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s

    Sign In Harbor  ${HARBOR_URL}  user021  Test1@34
    Create An New Project  project${d}
    Logout Harbor

    Manage Project Member  user021  Test1@34  project${d}  user022  Add  ${false}
    Change User Role In Project  user021  Test1@34  project${d}  user022  Developer

    Sign In Harbor  ${HARBOR_URL}  user022  Test1@34
    Go Into Project  project${d}  has_image=${false}
    Sleep  3
    Retry Wait Until Page Not Contains Element  xpath=//a[contains(.,'Labels')]
    Close Browser

Test Case - Retag A Image Tag
    Init Chrome Driver
    ${random_num1}=   Get Current Date    result_format=%m%s
    ${random_num2}=   Evaluate  str(random.randint(1000,9999))  modules=random

    Sign In Harbor  ${HARBOR_URL}  user028  Test1@34
    Create An New Project  project${random_num1}
    Create An New Project  project${random_num1}${random_num2}

    Go Into Project  project${random_num1}  has_image=${false}
    Sleep  1
    Push Image With Tag  ${ip}  user028  Test1@34  project${random_num1}  redis  ${image_tag}
    Sleep  1
    Go Into Repo  project${random_num1}/redis
    Retag Image  ${image_tag}  project${random_num1}${random_num2}  ${target_image_name}  ${target_tag_value}
    Retry Wait Element Not Visible  ${repo_retag_confirm_dlg}
    Navigate To Projects
    Go Into Project  project${random_num1}${random_num2}
    Sleep  1
    Page Should Contain  ${target_image_name}
    Go Into Repo  project${random_num1}${random_num2}/${target_image_name}
    Sleep  1
    Retry Wait Until Page Contains Element  xpath=${tag_value_xpath}
    Close Browser

Test Case - Create An New Project With Quotas Set
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${count_quota}=  Set Variable  1234
    ${storage_quota}=  Set Variable  600
    ${storage_quota_unit}=  Set Variable  GB
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project    project${d}  count_quota=${count_quota}  storage_quota=${storage_quota}  storage_quota_unit=${storage_quota_unit}
    ${count_quota_ret}=  Get Project Count Quota Text From Project Quotas List  project${d}
    Should Be Equal As Strings  ${count_quota_ret}  0 of ${count_quota}
    ${storage_quota_ret}=  Get Project Storage Quota Text From Project Quotas List  project${d}
    Should Be Equal As Strings  ${storage_quota_ret}  0Byte of ${storage_quota}${storage_quota_unit}
    Close Browser

Test Case - Project Image And Chart Artifact Count Quotas Dispaly And Control
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${count_quota}=  Set Variable  2
    ${storage_quota}=  Set Variable  500
    ${storage_quota_unit}=  Set Variable  MB
    ${image}=  Set Variable  redis
    #For docker-hub registry
    #${sha256}=  Set Variable  9755880356c4ced4ff7745bafe620f0b63dd17747caedba72504ef7bac882089
    #For internal CPE harbor registry
    ${sha256}=  Set Variable  0e67625224c1da47cb3270e7a861a83e332f708d3d89dde0cbed432c94824d9a

    ${image_size}=    Set Variable    34.14MB
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}  count_quota=${count_quota}  storage_quota=${storage_quota}  storage_quota_unit=${storage_quota_unit}
    Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image}  sha256=${sha256}
    ${count_quota_ret}=  Get Project Count Quota Text From Project Quotas List  project${d}
    Should Be Equal As Strings  ${count_quota_ret}  1 of ${count_quota}
    ${storage_quota_ret}=  Get Project Storage Quota Text From Project Quotas List  project${d}
    Should Be Equal As Strings  ${storage_quota_ret}  ${image_size} of ${storage_quota}${storage_quota_unit}
    #Push Image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  alpine
    Go Into Project  project${d}
    Switch To Project Charts
    Upload Chart files
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox  err_msg=Quota exceeded when processing the request of adding 1 of count resource, which when updated to current usage of 2 will exceed the configured upper limit of 2
    ${count_quota_ret}=  Get Project Count Quota Text From Project Quotas List  project${d}
    Should Be Equal As Strings  ${count_quota_ret}  2 of ${count_quota}
    Go Into Project  project${d}
    Delete Repo  project${d}/${image}
    ${count_quota_ret}=  Get Project Count Quota Text From Project Quotas List  project${d}
    Should Be Equal As Strings  ${count_quota_ret}  1 of ${count_quota}
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox
    Close Browser

Test Case - Project Storage Quotas Dispaly And Control
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${storage_quota}=  Set Variable  330
    ${storage_quota_unit}=  Set Variable  MB
    ${image_a}=  Set Variable  redis
    ${image_b}=  Set Variable  logstash
    ${image_a_size}=    Set Variable    34.15MB
    ${image_b_size}=    Set Variable    321.03MB
    ${image_a_ver}=  Set Variable  donotremove5.0
    ${image_b_ver}=  Set Variable  do_not_remove_6.8.3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}  storage_quota=${storage_quota}  storage_quota_unit=${storage_quota_unit}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_b}  tag=${image_b_ver}  tag1=${image_b_ver}
    ${storage_quota_ret}=  Get Project Storage Quota Text From Project Quotas List  project${d}
    Should Be Equal As Strings  ${storage_quota_ret}  ${image_b_size} of ${storage_quota}${storage_quota_unit}
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_a}:${image_a_ver}  err_msg=Quota exceeded when processing the request of adding 25.8 MiB of storage resource, which when updated to current usage of 329.3 MiB will exceed the configured upper limit of 330.0 MiB
    Go Into Project  project${d}
    Delete Repo  project${d}/${image_b}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_a}  tag=${image_a_ver}  tag1=${image_a_ver}
    ${storage_quota_ret}=  Get Project Storage Quota Text From Project Quotas List  project${d}
    ${storage_quota_ret_str_left}  Fetch From Left  ${storage_quota_ret}  25.
    Log  ${storage_quota_ret_str_left}
    ${storage_quota_ret_str_right}  Fetch From Left  ${storage_quota_ret}  25.
    Log  ${storage_quota_ret_str_right}
    Log  ${storage_quota_ret_str_left}${storage_quota_ret_str_right}
    Should Be Equal As Strings  ${storage_quota_ret}  ${image_a_size} of ${storage_quota}${storage_quota_unit}
    Close Browser

Test Case - Project Quotas Control Under Retag
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${count_quota}=  Set Variable  1
    ${image_a}=  Set Variable  redis
    ${image_b}=  Set Variable  logstash
    ${image_a_ver}=  Set Variable  5.0
    ${image_b_ver}=  Set Variable  6.8.3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project_a_${d}
    Create An New Project  project_b_${d}  count_quota=${count_quota}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project_a_${d}  ${image_a}  tag=${image_a_ver}  tag1=${image_a_ver}
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project_a_${d}  ${image_b}  tag=${image_b_ver}  tag1=${image_b_ver}
    Go Into Project  project_a_${d}
    Go Into Repo  project_a_${d}/${image_a}
    Retag Image  ${image_a_ver}  project_b_${d}  ${image_a}  ${image_a_ver}
    Retry Wait Element Not Visible  ${repo_retag_confirm_dlg}
    Go Into Project  project_a_${d}
    Go Into Repo  project_a_${d}/${image_b}
    Retag Image  ${image_b_ver}  project_b_${d}  ${image_b}  ${image_b_ver}
    Retry Wait Element Not Visible  ${repo_retag_confirm_dlg}
    Sleep  2
    Go Into Project  project_b_${d}
    Sleep  2
    Capture Page Screenshot
    Retry Wait Until Page Contains Element  xpath=//clr-dg-cell[contains(.,'${image_a}')]/a
    Retry Wait Until Page Not Contains Element  xpath=//clr-dg-cell[contains(.,'${image_b}')]/a
    Capture Page Screenshot
    Close Browser

Test Case - Project Quotas Control Under GC
    Init Chrome Driver
    ${d}=  Get Current Date  result_format=%m%s
    ${storage_quota}=  Set Variable  200
    ${storage_quota_unit}=  Set Variable  MB
    ${image_a}=  Set Variable  logstash
    ${image_a_size}=    Set Variable    321.03MB
    ${image_a_ver}=  Set Variable  6.8.3
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Capture Page Screenshot
    Create An New Project  project${d}  storage_quota=${storage_quota}  storage_quota_unit=${storage_quota_unit}
    Capture Page Screenshot
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  ${image_a}:${image_a_ver}  err_msg=will exceed the configured upper limit of 200.0 MiB
    Capture Page Screenshot
    GC Now  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    @{param}  Create List  project${d}
    Retry Keyword When Return Value Mismatch  Get Project Storage Quota Text From Project Quotas List  0Byte of ${storage_quota}${storage_quota_unit}  60  @{param}
    Close Browser

Test Case - Create New Webhook
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Switch To Project Webhooks
    Create A New Webhook  ${HARBOR_URL}  auth_header=auth_header${d}
    Close Browser

Test Case - Update Webhook
   Init Chrome Driver
   ${d}=    Get Current Date    result_format=%m%s
   Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
   Create An New Project  project${d}
   Go Into Project  project${d}  has_image=${false}
   Switch To Project Webhooks
   Create A New Webhook  ${HARBOR_URL}  auth_header=auth_header${d}
   Sleep  3
   ${d1}=    Get Current Date
   Update A Webhook  101.17.109.20  auth_header=auth_header${d1}
   Close Browser

Test Case - Toggle Enable/Disable State of Webhook
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Switch To Project Webhooks
    Create A New Webhook  ${HARBOR_URL}  auth_header=auth_header${d}
    Sleep  3
    Toggle Enable/Disable State of Same Webhook
    Close Browser

Test Case - Tag Retention
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=    Get Current Date    result_format=%m%s
    Create An New Project  project${d}
    Switch To Tag Retention
    Add A Tag Retention Rule
    Delete A Tag Retention Rule
    Add A Tag Retention Rule
    Edit A Tag Retention Rule    **   latest
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world  latest
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  memcached   123
    Set Daily Schedule
    Execute Dry Run
    Execute Run
    Close Browser

Test Case - Tag Immutability
    Init Chrome Driver
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    ${d}=    Get Current Date    result_format=%m%s
    Create An New Project  project${d}
    Go Into Project  project${d}  has_image=${false}
    Switch To Tag Immutability
    Add A Tag Immutability Rule  1212  3434
    Delete A Tag Immutability Rule
    Add A Tag Immutability Rule  5566  7788
    Edit A Tag Immutability Rule  hello-world  latest
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  hello-world  latest
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox  latest
    Go Into Project  project${d}
    @{repo_list}  Create List  hello-world  busybox
    Multi-delete Object  ${repo_delete_btn}  @{repo_list}
    # Verify
    Delete Fail  hello-world
    Delete Success  busybox
    Close Browser


Test Case - Robot Account
    Init Chrome Driver
    ${d}=    Get Current Date    result_format=%m%s
    Sign In Harbor    ${HARBOR_URL}    ${HARBOR_ADMIN}    ${HARBOR_PASSWORD}
    Create An New Project    project${d}
    ${token}=    Create A Robot Account And Return Token    project${d}    robot${d}
    Log To Console    ${token}
    Log    ${token}
    Push image  ${ip}  robot${d}  ${token}  project${d}  hello-world:latest  is_robot=${true}
    Pull image  ${ip}  robot${d}  ${token}  project${d}  hello-world:latest  is_robot=${true}

Test Case - Read Only Mode
    Init Chrome Driver
    ${d}=   Get Current Date    result_format=%m%s
    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${d}

    Enable Read Only
    Cannot Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest

    Disable Read Only
    Sleep  5
    Push image  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${d}  busybox:latest
    Close Browser

Test Case - Can Not Retag Image In ReadOnly Mode
    Init Chrome Driver
    ${random_num1}=   Get Current Date    result_format=%m%s
    ${random_num2}=   Evaluate  str(random.randint(1000,9999))  modules=random

    Sign In Harbor  ${HARBOR_URL}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}
    Create An New Project  project${random_num1}
    Create An New Project  project${random_num1}${random_num2}

    Go Into Project  project${random_num1}  has_image=${false}
    Sleep  1
    Push Image With Tag  ${ip}  ${HARBOR_ADMIN}  ${HARBOR_PASSWORD}  project${random_num1}  redis  ${image_tag}
    Sleep  1
    Enable Read Only
    Go Into Repo  project${random_num1}/redis
    Retag Image  ${image_tag}  project${random_num1}${random_num2}  ${target_image_name}  ${target_tag_value}
    Retry Wait Element Not Visible  ${repo_retag_confirm_dlg}
    Navigate To Projects
    Go Into Project  project${random_num1}${random_num2}  has_image=${false}
    Sleep  10
    Go Into Project  project${random_num1}${random_num2}  has_image=${false}
    Disable Read Only
    Close Browser


