// Copyright (c) 2017 VMware, Inc. All Rights Reserved.
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
import { debounceTime, distinctUntilChanged, filter, map, switchMap } from 'rxjs/operators';
import {
    Component,
    EventEmitter,
    Output,
    ViewChild,
    OnDestroy,
    Input,
    OnChanges,
    SimpleChanges, AfterViewInit, ElementRef
} from "@angular/core";
import { NgForm, Validators } from "@angular/forms";
import { fromEvent, Subscription } from "rxjs";
import { TranslateService } from "@ngx-translate/core";
import { MessageHandlerService } from "../../shared/message-handler/message-handler.service";
import { InlineAlertComponent } from "../../shared/inline-alert/inline-alert.component";
import { Project } from "../project";
import {
    clone, getByte,
    GetIntegerAndUnit,
    ProjectService,
    QuotaHardInterface,
    QuotaUnits,
    QuotaUnlimited, validateCountLimit,
    validateLimit
} from "@harbor/ui";



@Component({
  selector: "create-project",
  templateUrl: "create-project.component.html",
  styleUrls: ["create-project.scss"]
})
export class CreateProjectComponent implements  AfterViewInit, OnChanges, OnDestroy {

  projectForm: NgForm;

  @ViewChild("projectForm", {static: true})
  currentForm: NgForm;
  quotaUnits = QuotaUnits;
  project: Project = new Project();
  countLimit: number;
  storageLimit: number;
  storageLimitUnit: string = QuotaUnits[3].UNIT;
  storageDefaultLimit: number;
  storageDefaultLimitUnit: string;
  countDefaultLimit: number;
  initVal: Project = new Project();

  createProjectOpened: boolean;

  hasChanged: boolean;
  isSubmitOnGoing = false;

  staticBackdrop = true;
  closable = false;
  isNameExisted: boolean = false;
  nameTooltipText = "PROJECT.NAME_TOOLTIP";
  checkOnGoing = false;
  @Output() create = new EventEmitter<boolean>();
  @Input() quotaObj: QuotaHardInterface;
  @Input() isSystemAdmin: boolean;
  @ViewChild(InlineAlertComponent, {static: true})
  inlineAlert: InlineAlertComponent;
  @ViewChild('projectName', {static: false}) projectNameInput: ElementRef;
  checkNameSubscribe: Subscription;
  constructor(private projectService: ProjectService,
    private translateService: TranslateService,
    private messageHandlerService: MessageHandlerService) { }

    ngAfterViewInit(): void {
        if (!this.checkNameSubscribe) {
            this.checkNameSubscribe = fromEvent(this.projectNameInput.nativeElement, 'input').pipe(
                map((e: any) => e.target.value),
                debounceTime(300),
                distinctUntilChanged(),
                filter(name => {
                    return this.currentForm.controls["create_project_name"].valid && name.length > 0;
                }),
                switchMap(name => {
                    // Check exiting from backend
                    this.checkOnGoing = true;
                    this.isNameExisted = false;
                    return this.projectService.checkProjectExists(name);
                })).subscribe(response => {
                // Project existing
                if (!(response && response.status === 404)) {
                    this.isNameExisted = true;
                }
                this.checkOnGoing = false;
            }, error => {
                this.checkOnGoing = false;
                this.isNameExisted = false;
            });
        }
    }
   get isNameValid(): boolean {
        if (!this.currentForm || !this.currentForm.controls || !this.currentForm.controls["create_project_name"]) {
            return true;
        }
        if (!(this.currentForm.controls["create_project_name"].dirty || this.currentForm.controls["create_project_name"].touched)) {
            return true;
        }
        if (this.checkOnGoing) {
            return true;
        }
        if (this.currentForm.controls["create_project_name"].errors) {
            this.nameTooltipText = 'PROJECT.NAME_TOOLTIP';
            return false;
        }
        if (this.isNameExisted) {
            this.nameTooltipText = 'PROJECT.NAME_ALREADY_EXISTS';
            return false;
        }
        return true;
    }
  ngOnChanges(changes: SimpleChanges): void {
    if (changes && changes["quotaObj"] && changes["quotaObj"].currentValue) {
      this.countLimit = this.quotaObj.count_per_project;
      this.storageLimit = GetIntegerAndUnit(this.quotaObj.storage_per_project, clone(QuotaUnits), 0, clone(QuotaUnits)).partNumberHard;
      this.storageLimitUnit = this.storageLimit === QuotaUnlimited ? QuotaUnits[3].UNIT
      : GetIntegerAndUnit(this.quotaObj.storage_per_project, clone(QuotaUnits), 0, clone(QuotaUnits)).partCharacterHard;

      this.countDefaultLimit = this.countLimit;
      this.storageDefaultLimit = this.storageLimit;
      this.storageDefaultLimitUnit = this.storageLimitUnit;
      if (this.isSystemAdmin) {
        this.currentForm.form.controls['create_project_storage_limit'].setValidators(
          [
          Validators.required,
          Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
          validateLimit(this.currentForm.form.controls['create_project_storage_limit_unit'])
      ]);
      this.currentForm.form.controls['create_project_count_limit'].setValidators(
        [
          Validators.required,
          Validators.pattern('(^-1$)|(^([1-9]+)([0-9]+)*$)'),
          validateCountLimit()
        ]);
      }
      this.currentForm.form.valueChanges
            .pipe(distinctUntilChanged((a, b) => JSON.stringify(a) === JSON.stringify(b)))
            .subscribe((data) => {
              ['create_project_storage_limit', 'create_project_storage_limit_unit', 'create_project_count_limit'].forEach(fieldName => {
                if (this.currentForm.form.get(fieldName) && this.currentForm.form.get(fieldName).value !== null) {
                  this.currentForm.form.get(fieldName).updateValueAndValidity();
                }
              });
            });
    }
}
  ngOnDestroy(): void {
    if (this.checkNameSubscribe) {
        this.checkNameSubscribe.unsubscribe();
        this.checkNameSubscribe = null;
    }
  }

  onSubmit() {
    if (this.isSubmitOnGoing) {
      return ;
    }
    this.isSubmitOnGoing = true;
    const storageByte = +this.storageLimit === QuotaUnlimited ? this.storageLimit : getByte(+this.storageLimit, this.storageLimitUnit);
    this.projectService
      .createProject(this.project.name, this.project.metadata, +this.countLimit, +storageByte)
      .subscribe(
      status => {
        this.isSubmitOnGoing = false;

        this.create.emit(true);
        this.messageHandlerService.showSuccess("PROJECT.CREATED_SUCCESS");
        this.createProjectOpened = false;
      },
      error => {
        this.isSubmitOnGoing = false;
        this.inlineAlert.showInlineError(error);
      });
  }

  onCancel() {
      this.createProjectOpened = false;
  }

  newProject() {
    this.project = new Project();
    this.hasChanged = false;
    this.createProjectOpened = true;
    if (this.currentForm && this.currentForm.controls && this.currentForm.controls["create_project_name"]) {
        this.currentForm.controls["create_project_name"].reset();
    }
    this.inlineAlert.close();
    this.countLimit = this.countDefaultLimit ;
    this.storageLimit = this.storageDefaultLimit;
    this.storageLimitUnit = this.storageDefaultLimitUnit;
  }

  public get isValid(): boolean {
    return this.currentForm &&
    this.currentForm.valid &&
    !this.isSubmitOnGoing &&
    this.isNameValid &&
    !this.checkOnGoing;
  }
}

