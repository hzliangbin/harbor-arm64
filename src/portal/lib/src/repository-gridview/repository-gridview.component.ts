import {
    Component,
    Input,
    Output,
    OnInit,
    ViewChild,
    ChangeDetectionStrategy,
    ChangeDetectorRef,
    EventEmitter,
    OnChanges,
    SimpleChanges,
    Inject
} from "@angular/core";
import { Router } from "@angular/router";
import { forkJoin } from "rxjs";
import { finalize } from "rxjs/operators";
import { TranslateService } from "@ngx-translate/core";
import { Comparator, State } from "../service/interface";

import {
    Repository,
    SystemInfo,
    SystemInfoService,
    RepositoryService,
    RequestQueryParams,
    RepositoryItem,
    TagService
} from '../service/index';
import { ErrorHandler } from '../error-handler/error-handler';
import { CustomComparator, DEFAULT_PAGE_SIZE, calculatePage, doFiltering, doSorting, clone } from '../utils';
import { ConfirmationState, ConfirmationTargets, ConfirmationButtons } from '../shared/shared.const';
import { ConfirmationDialogComponent } from '../confirmation-dialog/confirmation-dialog.component';
import { ConfirmationMessage } from '../confirmation-dialog/confirmation-message';
import { ConfirmationAcknowledgement } from '../confirmation-dialog/confirmation-state-message';
import { Tag } from '../service/interface';
import { GridViewComponent } from '../gridview/grid-view.component';
import { OperationService } from "../operation/operation.service";
import { UserPermissionService } from "../service/permission.service";
import { USERSTATICPERMISSION } from "../service/permission-static";
import { OperateInfo, OperationState, operateChanges } from "../operation/operate";
import { SERVICE_CONFIG, IServiceConfig } from '../service.config';
import { map, catchError } from "rxjs/operators";
import { Observable, throwError as observableThrowError } from "rxjs";
import { errorHandler as errorHandFn } from "../shared/shared.utils";
@Component({
    selector: "hbr-repository-gridview",
    templateUrl: "./repository-gridview.component.html",
    styleUrls: ["./repository-gridview.component.scss"],
})
export class RepositoryGridviewComponent implements OnInit {
    signedCon: { [key: string]: any | string[] } = {};
    downloadLink: string;
    @Input() projectId: number;
    @Input() projectName = "unknown";
    @Input() urlPrefix: string;
    @Input() hasSignedIn: boolean;
    @Input() hasProjectAdminRole: boolean;
    @Input() mode = "admiral";
    @Output() repoClickEvent = new EventEmitter<RepositoryItem>();
    @Output() repoProvisionEvent = new EventEmitter<RepositoryItem>();
    @Output() addInfoEvent = new EventEmitter<RepositoryItem>();

    lastFilteredRepoName: string;
    repositories: RepositoryItem[] = [];
    repositoriesCopy: RepositoryItem[] = [];
    systemInfo: SystemInfo;
    selectedRow: RepositoryItem[] = [];
    loading = true;

    isCardView: boolean;
    cardHover = false;
    listHover = false;

    pullCountComparator: Comparator<RepositoryItem> = new CustomComparator<RepositoryItem>('pull_count', 'number');
    tagsCountComparator: Comparator<RepositoryItem> = new CustomComparator<RepositoryItem>('tags_count', 'number');

    pageSize: number = DEFAULT_PAGE_SIZE;
    currentPage = 1;
    totalCount = 0;
    currentState: State;

    @ViewChild("confirmationDialog", {static: false})
    confirmationDialog: ConfirmationDialogComponent;

    @ViewChild("gridView", {static: false}) gridView: GridViewComponent;
    hasCreateRepositoryPermission: boolean;
    hasDeleteRepositoryPermission: boolean;
    constructor(@Inject(SERVICE_CONFIG) private configInfo: IServiceConfig,
        private errorHandler: ErrorHandler,
        private translateService: TranslateService,
        private repositoryService: RepositoryService,
        private systemInfoService: SystemInfoService,
        private tagService: TagService,
        private operationService: OperationService,
        public userPermissionService: UserPermissionService,
        private router: Router) {
        if (this.configInfo && this.configInfo.systemInfoEndpoint) {
            this.downloadLink = this.configInfo.systemInfoEndpoint + "/getcert";
        }
    }

    public get registryUrl(): string {
        return this.systemInfo ? this.systemInfo.registry_url : "";
    }
    public get isClairDBReady(): boolean {
        return (
            this.systemInfo &&
            this.systemInfo.clair_vulnerability_status &&
            this.systemInfo.clair_vulnerability_status.overall_last_update > 0
        );
    }

    public get withAdmiral(): boolean {
        return this.mode === "admiral";
    }
    get canDownloadCert(): boolean {
        return this.systemInfo && this.systemInfo.has_ca_root;
    }
    ngOnInit(): void {
        // Get system info for tag views
        this.systemInfoService.getSystemInfo()
            .subscribe(systemInfo => (this.systemInfo = systemInfo)
                , error => this.errorHandler.error(error));
        this.isCardView = this.mode === "admiral";
        this.lastFilteredRepoName = "";
        this.getHelmChartVersionPermission(this.projectId);
    }

    confirmDeletion(message: ConfirmationAcknowledgement) {
        let repArr: any[] = [];
        message.data.forEach(repo => {
            if (!this.signedCon[repo.name]) {
                repArr.push(this.getTagInfo(repo.name));
            }
        });
        this.loading = true;
        forkJoin(...repArr).subscribe(() => {
            if (message &&
                message.source === ConfirmationTargets.REPOSITORY &&
                message.state === ConfirmationState.CONFIRMED) {
                let repoLists = message.data;
                if (repoLists && repoLists.length) {
                    let observableLists: any[] = [];
                    repoLists.forEach(repo => {
                        observableLists.push(this.delOperate(repo));
                    });
                    forkJoin(observableLists).subscribe((item) => {
                        this.selectedRow = [];
                        this.refresh();
                        let st: State = this.getStateAfterDeletion();
                        if (!st) {
                            this.refresh();
                        } else {
                            this.clrLoad(st);
                        }
                    }, error => {
                        this.errorHandler.error(error);
                        this.loading = false;
                        this.refresh();
                    });
                }
            }
        }, error => {
            this.errorHandler.error(error);
            this.loading = false;
        });
    }

    delOperate(repo: RepositoryItem): Observable<any> {
        // init operation info
        let operMessage = new OperateInfo();
        operMessage.name = 'OPERATION.DELETE_REPO';
        operMessage.data.id = repo.id;
        operMessage.state = OperationState.progressing;
        operMessage.data.name = repo.name;
        this.operationService.publishInfo(operMessage);

        if (this.signedCon[repo.name].length !== 0) {
            return forkJoin(this.translateService.get('BATCH.DELETED_FAILURE'),
                this.translateService.get('REPOSITORY.DELETION_TITLE_REPO_SIGNED')).pipe(map(res => {
                    operateChanges(operMessage, OperationState.failure, res[1]);
                }));
        } else {
            return this.repositoryService
                .deleteRepository(repo.name)
                .pipe(map(
                    response => {
                        this.translateService.get('BATCH.DELETED_SUCCESS').subscribe(res => {
                            operateChanges(operMessage, OperationState.success);
                        });
                    }), catchError(error => {
                        const message = errorHandFn(error);
                        this.translateService.get(message).subscribe(res =>
                            operateChanges(operMessage, OperationState.failure, res)
                        );
                        return observableThrowError(message);
                    }));
        }
    }

    doSearchRepoNames(repoName: string) {
        this.lastFilteredRepoName = repoName;
        this.currentPage = 1;
        let st: State = this.currentState;
        if (!st || !st.page) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = 0;
        st.page.to = this.pageSize - 1;
        this.clrLoad(st);
    }

    saveSignatures(event: { [key: string]: string[] }): void {
        Object.assign(this.signedCon, event);
    }

    deleteRepos(repoLists: RepositoryItem[]) {
        if (repoLists && repoLists.length) {
            let repoNames: string[] = [];
            repoLists.forEach(repo => {
                repoNames.push(repo.name);
            });
            this.confirmationDialogSet(
                'REPOSITORY.DELETION_TITLE_REPO',
                '',
                repoNames.join(','),
                repoLists,
                'REPOSITORY.DELETION_SUMMARY_REPO',
                ConfirmationButtons.DELETE_CANCEL);
        }
    }

    getTagInfo(repoName: string): Observable<void> {
        this.signedCon[repoName] = [];
        return this.tagService.getTags(repoName)
            .pipe(map(items => {
                items.forEach((t: Tag) => {
                    if (t.signature !== null) {
                        this.signedCon[repoName].push(t.name);
                    }
                });
            })
                , catchError(error => observableThrowError(error)));
    }

    confirmationDialogSet(summaryTitle: string, signature: string,
        repoName: string, repoLists: RepositoryItem[],
        summaryKey: string, button: ConfirmationButtons): void {
        this.translateService.get(summaryKey,
            {
                repoName: repoName,
                signedImages: signature,
            })
            .subscribe((res: string) => {
                summaryKey = res;
                let message = new ConfirmationMessage(
                    summaryTitle,
                    summaryKey,
                    repoName,
                    repoLists,
                    ConfirmationTargets.REPOSITORY,
                    button);
                this.confirmationDialog.open(message);
            });
    }

    containsLatestTag(repo: RepositoryItem): Observable<boolean> {
        return this.tagService.getTags(repo.name)
            .pipe(map(items => {
                if (items.some((t: Tag) => {
                    return t.name === 'latest';
                })) {
                    return true;
                } else {
                    return false;
                }

            })
                , catchError(error => observableThrowError(false)));
    }

    provisionItemEvent(evt: any, repo: RepositoryItem): void {
        evt.stopPropagation();
        let repoCopy = clone(repo);
        repoCopy.name = this.registryUrl + ":443/" + repoCopy.name;
        this.containsLatestTag(repo)
            .subscribe(containsLatest => {
                if (containsLatest) {
                    this.repoProvisionEvent.emit(repoCopy);
                } else {
                    this.addInfoEvent.emit(repoCopy);
                }
            }, error => this.errorHandler.error(error));

    }

    itemAddInfoEvent(evt: any, repo: RepositoryItem): void {
        evt.stopPropagation();
        let repoCopy = clone(repo);
        repoCopy.name = this.registryUrl + ":443/" + repoCopy.name;
        this.addInfoEvent.emit(repoCopy);
    }

    deleteItemEvent(evt: any, item: RepositoryItem): void {
        evt.stopPropagation();
        this.deleteRepos([item]);
    }
    refresh() {
        this.doSearchRepoNames("");
    }

    loadNextPage() {
        this.currentPage = this.currentPage + 1;
        // Pagination
        let params: RequestQueryParams = new RequestQueryParams().set("page", "" + this.currentPage).set("page_size", "" + this.pageSize);

        this.loading = true;
        this.repositoryService.getRepositories(
            this.projectId,
            this.lastFilteredRepoName,
            params
        ).pipe(finalize(() => this.loading = false))
            .subscribe((repo: Repository) => {
                this.totalCount = repo.metadata.xTotalCount;
                this.repositoriesCopy = repo.data;
                this.signedCon = {};
                // Do filtering and sorting
                this.repositoriesCopy = doFiltering<RepositoryItem>(
                    this.repositoriesCopy,
                    this.currentState
                );
                this.repositoriesCopy = doSorting<RepositoryItem>(
                    this.repositoriesCopy,
                    this.currentState
                );
                this.repositories = this.repositories.concat(this.repositoriesCopy);
            }, error => {
                this.errorHandler.error(error);
            });
    }

    clrLoad(state: State): void {
        if (!state || !state.page) {
            return;
        }
        this.selectedRow = [];
        // Keep it for future filtering and sorting
        this.currentState = state;

        let pageNumber: number = calculatePage(state);
        if (pageNumber <= 0) {
            pageNumber = 1;
        }

        // Pagination
        let params: RequestQueryParams = new RequestQueryParams().set("page", "" + pageNumber).set("page_size", "" + this.pageSize);
        this.loading = true;
        this.repositoryService.getRepositories(
            this.projectId,
            this.lastFilteredRepoName,
            params
        ).pipe(finalize(() => this.loading = false))
            .subscribe((repo: Repository) => {

                this.totalCount = repo.metadata.xTotalCount;
                this.repositories = repo.data;

                this.signedCon = {};
                // Do filtering and sorting
                this.repositories = doFiltering<RepositoryItem>(
                    this.repositories,
                    state
                );
                this.repositories = doSorting<RepositoryItem>(this.repositories, state);
            }, error => {
                this.errorHandler.error(error);
            });
    }

    getStateAfterDeletion(): State {
        let total: number = this.totalCount - 1;
        if (total <= 0) {
            return null;
        }

        let totalPages: number = Math.ceil(total / this.pageSize);
        let targetPageNumber: number = this.currentPage;

        if (this.currentPage > totalPages) {
            targetPageNumber = totalPages; // Should == currentPage -1
        }

        let st: State = this.currentState;
        if (!st) {
            st = { page: {} };
        }
        st.page.size = this.pageSize;
        st.page.from = (targetPageNumber - 1) * this.pageSize;
        st.page.to = targetPageNumber * this.pageSize - 1;

        return st;
    }

    watchRepoClickEvt(repo: RepositoryItem) {
        this.repoClickEvent.emit(repo);
    }

    getImgLink(repo: RepositoryItem): string {
        return "/container-image-icons?container-image=" + repo.name;
    }

    showCard(cardView: boolean) {
        if (this.isCardView === cardView) {
            return;
        }
        this.isCardView = cardView;
        this.refresh();
    }

    mouseEnter(itemName: string) {
        if (itemName === "card") {
            this.cardHover = true;
        } else {
            this.listHover = true;
        }
    }

    mouseLeave(itemName: string) {
        if (itemName === "card") {
            this.cardHover = false;
        } else {
            this.listHover = false;
        }
    }

    isHovering(itemName: string) {
        if (itemName === "card") {
            return this.cardHover;
        } else {
            return this.listHover;
        }
    }

    getHelmChartVersionPermission(projectId: number): void {

        let hasCreateRepositoryPermission = this.userPermissionService.getPermission(this.projectId,
            USERSTATICPERMISSION.REPOSITORY.KEY, USERSTATICPERMISSION.REPOSITORY.VALUE.CREATE);
        let hasDeleteRepositoryPermission = this.userPermissionService.getPermission(this.projectId,
            USERSTATICPERMISSION.REPOSITORY.KEY, USERSTATICPERMISSION.REPOSITORY.VALUE.DELETE);
        forkJoin(hasCreateRepositoryPermission, hasDeleteRepositoryPermission).subscribe(permissions => {
            this.hasCreateRepositoryPermission = permissions[0] as boolean;
            this.hasDeleteRepositoryPermission = permissions[1] as boolean;
        }, error => this.errorHandler.error(error));
    }
}
