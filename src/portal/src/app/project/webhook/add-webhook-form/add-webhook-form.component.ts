import {
  Component,
  OnInit,
  OnChanges,
  Input,
  ViewChild,
  Output,
  EventEmitter,
  SimpleChanges
} from "@angular/core";
import { Webhook, Target } from "../webhook";
import { NgForm } from "@angular/forms";
import { ClrLoadingState } from "@clr/angular";
import { finalize } from "rxjs/operators";
import { WebhookService } from "../webhook.service";
import { WebhookEventTypes } from '../../../shared/shared.const';
import { InlineAlertComponent } from "../../../shared/inline-alert/inline-alert.component";
import { MessageHandlerService } from "../../../shared/message-handler/message-handler.service";
import { TranslateService } from "@ngx-translate/core";

@Component({
  selector: 'add-webhook-form',
  templateUrl: './add-webhook-form.component.html',
  styleUrls: ['./add-webhook-form.component.scss']
})
export class AddWebhookFormComponent implements OnInit, OnChanges {
  closable: boolean = true;
  staticBackdrop: boolean = true;
  checking: boolean = false;
  checkBtnState: ClrLoadingState = ClrLoadingState.DEFAULT;
  webhookForm: NgForm;
  submitting: boolean = false;
  webhookTarget: Target = new Target();

  @Input() projectId: number;
  @Input() webhook: Webhook;
  @Input() isModify: boolean;
  @Input() isOpen: boolean;
  @Output() edit = new EventEmitter<boolean>();
  @Output() close = new EventEmitter<boolean>();
  @ViewChild("webhookForm", { static: true }) currentForm: NgForm;
  @ViewChild(InlineAlertComponent, { static: false }) inlineAlert: InlineAlertComponent;
  hasCreatPermission: boolean = false;
  hasUpdatePermission: boolean = false;

  constructor(
    private webhookService: WebhookService,
    private messageHandlerService: MessageHandlerService,
    private translate: TranslateService
  ) { }

  ngOnInit() {
   this.getPermissions();
  }
  getPermissions() {
    this.webhookService.getPermissions(this.projectId).subscribe(
      rules => {
        [this.hasCreatPermission, this.hasUpdatePermission] = rules;
      }
    );
  }

  ngOnChanges(changes: SimpleChanges) {
    if (changes['isOpen'] && changes['isOpen'].currentValue) {
      Object.assign(this.webhookTarget, this.webhook.targets[0]);
    }
  }
  onTestEndpoint() {
    this.checkBtnState = ClrLoadingState.LOADING;
    this.checking = true;

    this.webhookService
      .testEndpoint(this.projectId, {
        targets: [this.webhookTarget]
      })
      .pipe(finalize(() => (this.checking = false)))
      .subscribe(
        response => {
          if (this.isModify) {
            this.inlineAlert.showInlineSuccess({
              message: "WEBHOOK.TEST_ENDPOINT_SUCCESS"
            });
          } else {
            this.translate.get("WEBHOOK.TEST_ENDPOINT_SUCCESS").subscribe((res: string) => {
              this.messageHandlerService.info(res);
            });
          }
          this.checkBtnState = ClrLoadingState.SUCCESS;
        },
        error => {
          if (this.isModify) {
            this.inlineAlert.showInlineError("WEBHOOK.TEST_ENDPOINT_FAILURE");
          } else {
            this.messageHandlerService.handleError(error);
          }
          this.checkBtnState = ClrLoadingState.DEFAULT;
        }
      );
  }

  onCancel() {
    this.close.emit(false);
    this.currentForm.reset();
    this.inlineAlert.close();
  }

  onSubmit() {
    const rx = this.isModify
      ? this.webhookService.editWebhook(this.projectId, this.webhook.id, Object.assign(this.webhook, { targets: [this.webhookTarget] }))
      : this.webhookService.createWebhook(this.projectId, {
        targets: [this.webhookTarget],
        event_types: Object.keys(WebhookEventTypes).map(key => WebhookEventTypes[key]),
        enabled: true,
      });
    rx.pipe(finalize(() => (this.submitting = false)))
      .subscribe(
        response => {
          this.edit.emit(this.isModify);
          this.inlineAlert.close();
        },
        error => {
          this.isModify
            ? this.inlineAlert.showInlineError(error)
            : this.messageHandlerService.handleError(error);
        }
      );
  }

  setCertValue($event: any): void {
    this.webhookTarget.skip_cert_verify = !$event;
  }

  public get isValid(): boolean {
    return (
      this.currentForm &&
      this.currentForm.valid &&
      !this.submitting &&
      !this.checking
    );
  }
}
