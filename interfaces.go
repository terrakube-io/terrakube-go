package terrakube

import "context"

// OrganizationLister lists organizations.
type OrganizationLister interface {
	List(ctx context.Context, opts *ListOptions) ([]*Organization, error)
}

// OrganizationGetter retrieves a single organization.
type OrganizationGetter interface {
	Get(ctx context.Context, id string) (*Organization, error)
}

// OrganizationCreator creates organizations.
type OrganizationCreator interface {
	Create(ctx context.Context, org *Organization) (*Organization, error)
}

// OrganizationUpdater updates organizations.
type OrganizationUpdater interface {
	Update(ctx context.Context, org *Organization) (*Organization, error)
}

// OrganizationDeleter deletes organizations.
type OrganizationDeleter interface {
	Delete(ctx context.Context, id string) error
}

// OrganizationCRUD combines all organization operations.
type OrganizationCRUD interface {
	OrganizationLister
	OrganizationGetter
	OrganizationCreator
	OrganizationUpdater
	OrganizationDeleter
}

// WorkspaceLister lists workspaces.
type WorkspaceLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Workspace, error)
}

// WorkspaceGetter retrieves a single workspace.
type WorkspaceGetter interface {
	Get(ctx context.Context, orgID, id string) (*Workspace, error)
}

// WorkspaceCreator creates workspaces.
type WorkspaceCreator interface {
	Create(ctx context.Context, orgID string, ws *Workspace) (*Workspace, error)
}

// WorkspaceUpdater updates workspaces.
type WorkspaceUpdater interface {
	Update(ctx context.Context, orgID string, ws *Workspace) (*Workspace, error)
}

// WorkspaceDeleter deletes workspaces.
type WorkspaceDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// WorkspaceCRUD combines all workspace operations.
type WorkspaceCRUD interface {
	WorkspaceLister
	WorkspaceGetter
	WorkspaceCreator
	WorkspaceUpdater
	WorkspaceDeleter
}

// ModuleLister lists modules.
type ModuleLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Module, error)
}

// ModuleGetter retrieves a single module.
type ModuleGetter interface {
	Get(ctx context.Context, orgID, id string) (*Module, error)
}

// ModuleCreator creates modules.
type ModuleCreator interface {
	Create(ctx context.Context, orgID string, module *Module) (*Module, error)
}

// ModuleUpdater updates modules.
type ModuleUpdater interface {
	Update(ctx context.Context, orgID string, module *Module) (*Module, error)
}

// ModuleDeleter deletes modules.
type ModuleDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// ModuleCRUD combines all module operations.
type ModuleCRUD interface {
	ModuleLister
	ModuleGetter
	ModuleCreator
	ModuleUpdater
	ModuleDeleter
}

// TeamLister lists teams.
type TeamLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Team, error)
}

// TeamGetter retrieves a single team.
type TeamGetter interface {
	Get(ctx context.Context, orgID, id string) (*Team, error)
}

// TeamCreator creates teams.
type TeamCreator interface {
	Create(ctx context.Context, orgID string, team *Team) (*Team, error)
}

// TeamUpdater updates teams.
type TeamUpdater interface {
	Update(ctx context.Context, orgID string, team *Team) (*Team, error)
}

// TeamDeleter deletes teams.
type TeamDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// TeamCRUD combines all team operations.
type TeamCRUD interface {
	TeamLister
	TeamGetter
	TeamCreator
	TeamUpdater
	TeamDeleter
}

// TeamTokenLister lists team tokens.
type TeamTokenLister interface {
	List(ctx context.Context) ([]TeamToken, error)
}

// TeamTokenCreator creates team tokens.
type TeamTokenCreator interface {
	Create(ctx context.Context, token *TeamToken) (*TeamToken, error)
}

// TeamTokenDeleter deletes team tokens.
type TeamTokenDeleter interface {
	Delete(ctx context.Context, id string) error
}

// TeamTokenCRUD combines all team token operations.
type TeamTokenCRUD interface {
	TeamTokenLister
	TeamTokenCreator
	TeamTokenDeleter
}

// VariableLister lists variables.
type VariableLister interface {
	List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*Variable, error)
}

// VariableGetter retrieves a single variable.
type VariableGetter interface {
	Get(ctx context.Context, orgID, workspaceID, id string) (*Variable, error)
}

// VariableCreator creates variables.
type VariableCreator interface {
	Create(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error)
}

// VariableUpdater updates variables.
type VariableUpdater interface {
	Update(ctx context.Context, orgID, workspaceID string, variable *Variable) (*Variable, error)
}

// VariableDeleter deletes variables.
type VariableDeleter interface {
	Delete(ctx context.Context, orgID, workspaceID, id string) error
}

// VariableCRUD combines all variable operations.
type VariableCRUD interface {
	VariableLister
	VariableGetter
	VariableCreator
	VariableUpdater
	VariableDeleter
}

// OrganizationVariableLister lists organization variables.
type OrganizationVariableLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*OrganizationVariable, error)
}

// OrganizationVariableGetter retrieves a single organization variable.
type OrganizationVariableGetter interface {
	Get(ctx context.Context, orgID, id string) (*OrganizationVariable, error)
}

// OrganizationVariableCreator creates organization variables.
type OrganizationVariableCreator interface {
	Create(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error)
}

// OrganizationVariableUpdater updates organization variables.
type OrganizationVariableUpdater interface {
	Update(ctx context.Context, orgID string, variable *OrganizationVariable) (*OrganizationVariable, error)
}

// OrganizationVariableDeleter deletes organization variables.
type OrganizationVariableDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// OrganizationVariableCRUD combines all organization variable operations.
type OrganizationVariableCRUD interface {
	OrganizationVariableLister
	OrganizationVariableGetter
	OrganizationVariableCreator
	OrganizationVariableUpdater
	OrganizationVariableDeleter
}

// TemplateLister lists templates.
type TemplateLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Template, error)
}

// TemplateGetter retrieves a single template.
type TemplateGetter interface {
	Get(ctx context.Context, orgID, id string) (*Template, error)
}

// TemplateCreator creates templates.
type TemplateCreator interface {
	Create(ctx context.Context, orgID string, tmpl *Template) (*Template, error)
}

// TemplateUpdater updates templates.
type TemplateUpdater interface {
	Update(ctx context.Context, orgID string, tmpl *Template) (*Template, error)
}

// TemplateDeleter deletes templates.
type TemplateDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// TemplateCRUD combines all template operations.
type TemplateCRUD interface {
	TemplateLister
	TemplateGetter
	TemplateCreator
	TemplateUpdater
	TemplateDeleter
}

// TagLister lists tags.
type TagLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Tag, error)
}

// TagGetter retrieves a single tag.
type TagGetter interface {
	Get(ctx context.Context, orgID, id string) (*Tag, error)
}

// TagCreator creates tags.
type TagCreator interface {
	Create(ctx context.Context, orgID string, tag *Tag) (*Tag, error)
}

// TagUpdater updates tags.
type TagUpdater interface {
	Update(ctx context.Context, orgID string, tag *Tag) (*Tag, error)
}

// TagDeleter deletes tags.
type TagDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// TagCRUD combines all tag operations.
type TagCRUD interface {
	TagLister
	TagGetter
	TagCreator
	TagUpdater
	TagDeleter
}

// VCSLister lists VCS connections.
type VCSLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*VCS, error)
}

// VCSGetter retrieves a single VCS connection.
type VCSGetter interface {
	Get(ctx context.Context, orgID, id string) (*VCS, error)
}

// VCSCreator creates VCS connections.
type VCSCreator interface {
	Create(ctx context.Context, orgID string, vcs *VCS) (*VCS, error)
}

// VCSUpdater updates VCS connections.
type VCSUpdater interface {
	Update(ctx context.Context, orgID string, vcs *VCS) (*VCS, error)
}

// VCSDeleter deletes VCS connections.
type VCSDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// VCSCRUD combines all VCS operations.
type VCSCRUD interface {
	VCSLister
	VCSGetter
	VCSCreator
	VCSUpdater
	VCSDeleter
}

// SSHLister lists SSH keys.
type SSHLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*SSH, error)
}

// SSHGetter retrieves a single SSH key.
type SSHGetter interface {
	Get(ctx context.Context, orgID, id string) (*SSH, error)
}

// SSHCreator creates SSH keys.
type SSHCreator interface {
	Create(ctx context.Context, orgID string, ssh *SSH) (*SSH, error)
}

// SSHUpdater updates SSH keys.
type SSHUpdater interface {
	Update(ctx context.Context, orgID string, ssh *SSH) (*SSH, error)
}

// SSHDeleter deletes SSH keys.
type SSHDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// SSHCRUD combines all SSH operations.
type SSHCRUD interface {
	SSHLister
	SSHGetter
	SSHCreator
	SSHUpdater
	SSHDeleter
}

// AgentLister lists agents.
type AgentLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Agent, error)
}

// AgentGetter retrieves a single agent.
type AgentGetter interface {
	Get(ctx context.Context, orgID, id string) (*Agent, error)
}

// AgentCreator creates agents.
type AgentCreator interface {
	Create(ctx context.Context, orgID string, agent *Agent) (*Agent, error)
}

// AgentUpdater updates agents.
type AgentUpdater interface {
	Update(ctx context.Context, orgID string, agent *Agent) (*Agent, error)
}

// AgentDeleter deletes agents.
type AgentDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// AgentCRUD combines all agent operations.
type AgentCRUD interface {
	AgentLister
	AgentGetter
	AgentCreator
	AgentUpdater
	AgentDeleter
}

// CollectionLister lists collections.
type CollectionLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Collection, error)
}

// CollectionGetter retrieves a single collection.
type CollectionGetter interface {
	Get(ctx context.Context, orgID, id string) (*Collection, error)
}

// CollectionCreator creates collections.
type CollectionCreator interface {
	Create(ctx context.Context, orgID string, collection *Collection) (*Collection, error)
}

// CollectionUpdater updates collections.
type CollectionUpdater interface {
	Update(ctx context.Context, orgID string, collection *Collection) (*Collection, error)
}

// CollectionDeleter deletes collections.
type CollectionDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// CollectionCRUD combines all collection operations.
type CollectionCRUD interface {
	CollectionLister
	CollectionGetter
	CollectionCreator
	CollectionUpdater
	CollectionDeleter
}

// CollectionItemLister lists collection items.
type CollectionItemLister interface {
	List(ctx context.Context, orgID, collectionID string, opts *ListOptions) ([]*CollectionItem, error)
}

// CollectionItemGetter retrieves a single collection item.
type CollectionItemGetter interface {
	Get(ctx context.Context, orgID, collectionID, id string) (*CollectionItem, error)
}

// CollectionItemCreator creates collection items.
type CollectionItemCreator interface {
	Create(ctx context.Context, orgID, collectionID string, item *CollectionItem) (*CollectionItem, error)
}

// CollectionItemUpdater updates collection items.
type CollectionItemUpdater interface {
	Update(ctx context.Context, orgID, collectionID string, item *CollectionItem) (*CollectionItem, error)
}

// CollectionItemDeleter deletes collection items.
type CollectionItemDeleter interface {
	Delete(ctx context.Context, orgID, collectionID, id string) error
}

// CollectionItemCRUD combines all collection item operations.
type CollectionItemCRUD interface {
	CollectionItemLister
	CollectionItemGetter
	CollectionItemCreator
	CollectionItemUpdater
	CollectionItemDeleter
}

// CollectionReferenceLister lists collection references.
type CollectionReferenceLister interface {
	List(ctx context.Context, orgID, collectionID string, opts *ListOptions) ([]*CollectionReference, error)
}

// CollectionReferenceGetter retrieves a single collection reference.
type CollectionReferenceGetter interface {
	Get(ctx context.Context, id string) (*CollectionReference, error)
}

// CollectionReferenceCreator creates collection references.
type CollectionReferenceCreator interface {
	Create(ctx context.Context, orgID, collectionID string, ref *CollectionReference) (*CollectionReference, error)
}

// CollectionReferenceUpdater updates collection references.
type CollectionReferenceUpdater interface {
	Update(ctx context.Context, ref *CollectionReference) (*CollectionReference, error)
}

// CollectionReferenceDeleter deletes collection references.
type CollectionReferenceDeleter interface {
	Delete(ctx context.Context, id string) error
}

// CollectionReferenceCRUD combines all collection reference operations.
type CollectionReferenceCRUD interface {
	CollectionReferenceLister
	CollectionReferenceGetter
	CollectionReferenceCreator
	CollectionReferenceUpdater
	CollectionReferenceDeleter
}

// WorkspaceTagLister lists workspace tags.
type WorkspaceTagLister interface {
	List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*WorkspaceTag, error)
}

// WorkspaceTagGetter retrieves a single workspace tag.
type WorkspaceTagGetter interface {
	Get(ctx context.Context, orgID, workspaceID, tagID string) (*WorkspaceTag, error)
}

// WorkspaceTagCreator creates workspace tags.
type WorkspaceTagCreator interface {
	Create(ctx context.Context, orgID, workspaceID string, tag *WorkspaceTag) (*WorkspaceTag, error)
}

// WorkspaceTagUpdater updates workspace tags.
type WorkspaceTagUpdater interface {
	Update(ctx context.Context, orgID, workspaceID string, tag *WorkspaceTag) (*WorkspaceTag, error)
}

// WorkspaceTagDeleter deletes workspace tags.
type WorkspaceTagDeleter interface {
	Delete(ctx context.Context, orgID, workspaceID, tagID string) error
}

// WorkspaceTagCRUD combines all workspace tag operations.
type WorkspaceTagCRUD interface {
	WorkspaceTagLister
	WorkspaceTagGetter
	WorkspaceTagCreator
	WorkspaceTagUpdater
	WorkspaceTagDeleter
}

// WorkspaceAccessLister lists workspace access entries.
type WorkspaceAccessLister interface {
	List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*WorkspaceAccess, error)
}

// WorkspaceAccessGetter retrieves a single workspace access entry.
type WorkspaceAccessGetter interface {
	Get(ctx context.Context, orgID, workspaceID, id string) (*WorkspaceAccess, error)
}

// WorkspaceAccessCreator creates workspace access entries.
type WorkspaceAccessCreator interface {
	Create(ctx context.Context, orgID, workspaceID string, access *WorkspaceAccess) (*WorkspaceAccess, error)
}

// WorkspaceAccessUpdater updates workspace access entries.
type WorkspaceAccessUpdater interface {
	Update(ctx context.Context, orgID, workspaceID string, access *WorkspaceAccess) (*WorkspaceAccess, error)
}

// WorkspaceAccessDeleter deletes workspace access entries.
type WorkspaceAccessDeleter interface {
	Delete(ctx context.Context, orgID, workspaceID, id string) error
}

// WorkspaceAccessCRUD combines all workspace access operations.
type WorkspaceAccessCRUD interface {
	WorkspaceAccessLister
	WorkspaceAccessGetter
	WorkspaceAccessCreator
	WorkspaceAccessUpdater
	WorkspaceAccessDeleter
}

// WorkspaceScheduleLister lists workspace schedules.
type WorkspaceScheduleLister interface {
	List(ctx context.Context, workspaceID string, opts *ListOptions) ([]*WorkspaceSchedule, error)
}

// WorkspaceScheduleGetter retrieves a single workspace schedule.
type WorkspaceScheduleGetter interface {
	Get(ctx context.Context, workspaceID, id string) (*WorkspaceSchedule, error)
}

// WorkspaceScheduleCreator creates workspace schedules.
type WorkspaceScheduleCreator interface {
	Create(ctx context.Context, workspaceID string, schedule *WorkspaceSchedule) (*WorkspaceSchedule, error)
}

// WorkspaceScheduleUpdater updates workspace schedules.
type WorkspaceScheduleUpdater interface {
	Update(ctx context.Context, workspaceID string, schedule *WorkspaceSchedule) (*WorkspaceSchedule, error)
}

// WorkspaceScheduleDeleter deletes workspace schedules.
type WorkspaceScheduleDeleter interface {
	Delete(ctx context.Context, workspaceID, id string) error
}

// WorkspaceScheduleCRUD combines all workspace schedule operations.
type WorkspaceScheduleCRUD interface {
	WorkspaceScheduleLister
	WorkspaceScheduleGetter
	WorkspaceScheduleCreator
	WorkspaceScheduleUpdater
	WorkspaceScheduleDeleter
}

// WebhookLister lists webhooks.
type WebhookLister interface {
	List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*Webhook, error)
}

// WebhookGetter retrieves a single webhook.
type WebhookGetter interface {
	Get(ctx context.Context, orgID, workspaceID, webhookID string) (*Webhook, error)
}

// WebhookCreator creates webhooks.
type WebhookCreator interface {
	Create(ctx context.Context, orgID, workspaceID string, webhook *Webhook) (*Webhook, error)
}

// WebhookUpdater updates webhooks.
type WebhookUpdater interface {
	Update(ctx context.Context, orgID, workspaceID string, webhook *Webhook) (*Webhook, error)
}

// WebhookDeleter deletes webhooks.
type WebhookDeleter interface {
	Delete(ctx context.Context, orgID, workspaceID, webhookID string) error
}

// WebhookCRUD combines all webhook operations.
type WebhookCRUD interface {
	WebhookLister
	WebhookGetter
	WebhookCreator
	WebhookUpdater
	WebhookDeleter
}

// WebhookEventLister lists webhook events.
type WebhookEventLister interface {
	List(ctx context.Context, orgID, workspaceID, webhookID string, opts *ListOptions) ([]*WebhookEvent, error)
}

// WebhookEventGetter retrieves a single webhook event.
type WebhookEventGetter interface {
	Get(ctx context.Context, orgID, workspaceID, webhookID, eventID string) (*WebhookEvent, error)
}

// WebhookEventCreator creates webhook events.
type WebhookEventCreator interface {
	Create(ctx context.Context, orgID, workspaceID, webhookID string, event *WebhookEvent) (*WebhookEvent, error)
}

// WebhookEventUpdater updates webhook events.
type WebhookEventUpdater interface {
	Update(ctx context.Context, orgID, workspaceID, webhookID string, event *WebhookEvent) (*WebhookEvent, error)
}

// WebhookEventDeleter deletes webhook events.
type WebhookEventDeleter interface {
	Delete(ctx context.Context, orgID, workspaceID, webhookID, eventID string) error
}

// WebhookEventCRUD combines all webhook event operations.
type WebhookEventCRUD interface {
	WebhookEventLister
	WebhookEventGetter
	WebhookEventCreator
	WebhookEventUpdater
	WebhookEventDeleter
}

// HistoryLister lists history entries.
type HistoryLister interface {
	List(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*History, error)
}

// HistoryGetter retrieves a single history entry.
type HistoryGetter interface {
	Get(ctx context.Context, orgID, workspaceID, id string) (*History, error)
}

// HistoryCreator creates history entries.
type HistoryCreator interface {
	Create(ctx context.Context, orgID, workspaceID string, history *History) (*History, error)
}

// HistoryUpdater updates history entries.
type HistoryUpdater interface {
	Update(ctx context.Context, orgID, workspaceID string, history *History) (*History, error)
}

// HistoryDeleter deletes history entries.
type HistoryDeleter interface {
	Delete(ctx context.Context, orgID, workspaceID, id string) error
}

// HistoryCRUD combines all history operations.
type HistoryCRUD interface {
	HistoryLister
	HistoryGetter
	HistoryCreator
	HistoryUpdater
	HistoryDeleter
}

// JobLister lists jobs.
type JobLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Job, error)
}

// JobGetter retrieves a single job.
type JobGetter interface {
	Get(ctx context.Context, orgID, id string) (*Job, error)
}

// JobCreator creates jobs.
type JobCreator interface {
	Create(ctx context.Context, orgID string, job *Job) (*Job, error)
}

// JobUpdater updates jobs.
type JobUpdater interface {
	Update(ctx context.Context, orgID string, job *Job) (*Job, error)
}

// JobDeleter deletes jobs.
type JobDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// JobCRUD combines all job operations.
type JobCRUD interface {
	JobLister
	JobGetter
	JobCreator
	JobUpdater
	JobDeleter
}

// ActionLister lists actions.
type ActionLister interface {
	List(ctx context.Context, opts *ListOptions) ([]*Action, error)
}

// ActionGetter retrieves a single action.
type ActionGetter interface {
	Get(ctx context.Context, id string) (*Action, error)
}

// ActionCreator creates actions.
type ActionCreator interface {
	Create(ctx context.Context, action *Action) (*Action, error)
}

// ActionUpdater updates actions.
type ActionUpdater interface {
	Update(ctx context.Context, action *Action) (*Action, error)
}

// ActionDeleter deletes actions.
type ActionDeleter interface {
	Delete(ctx context.Context, id string) error
}

// ActionCRUD combines all action operations.
type ActionCRUD interface {
	ActionLister
	ActionGetter
	ActionCreator
	ActionUpdater
	ActionDeleter
}

// StepLister lists steps.
type StepLister interface {
	List(ctx context.Context, orgID, jobID string, opts *ListOptions) ([]*Step, error)
}

// StepGetter retrieves a single step.
type StepGetter interface {
	Get(ctx context.Context, orgID, jobID, id string) (*Step, error)
}

// StepCreator creates steps.
type StepCreator interface {
	Create(ctx context.Context, orgID, jobID string, step *Step) (*Step, error)
}

// StepUpdater updates steps.
type StepUpdater interface {
	Update(ctx context.Context, orgID, jobID string, step *Step) (*Step, error)
}

// StepDeleter deletes steps.
type StepDeleter interface {
	Delete(ctx context.Context, orgID, jobID, id string) error
}

// StepCRUD combines all step operations.
type StepCRUD interface {
	StepLister
	StepGetter
	StepCreator
	StepUpdater
	StepDeleter
}

// ProviderLister lists providers.
type ProviderLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Provider, error)
}

// ProviderGetter retrieves a single provider.
type ProviderGetter interface {
	Get(ctx context.Context, orgID, id string) (*Provider, error)
}

// ProviderCreator creates providers.
type ProviderCreator interface {
	Create(ctx context.Context, orgID string, provider *Provider) (*Provider, error)
}

// ProviderUpdater updates providers.
type ProviderUpdater interface {
	Update(ctx context.Context, orgID string, provider *Provider) (*Provider, error)
}

// ProviderDeleter deletes providers.
type ProviderDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// ProviderCRUD combines all provider operations.
type ProviderCRUD interface {
	ProviderLister
	ProviderGetter
	ProviderCreator
	ProviderUpdater
	ProviderDeleter
}

// ProviderVersionLister lists provider versions.
type ProviderVersionLister interface {
	List(ctx context.Context, orgID, providerID string, opts *ListOptions) ([]*ProviderVersion, error)
}

// ProviderVersionGetter retrieves a single provider version.
type ProviderVersionGetter interface {
	Get(ctx context.Context, orgID, providerID, id string) (*ProviderVersion, error)
}

// ProviderVersionCreator creates provider versions.
type ProviderVersionCreator interface {
	Create(ctx context.Context, orgID, providerID string, version *ProviderVersion) (*ProviderVersion, error)
}

// ProviderVersionUpdater updates provider versions.
type ProviderVersionUpdater interface {
	Update(ctx context.Context, orgID, providerID string, version *ProviderVersion) (*ProviderVersion, error)
}

// ProviderVersionDeleter deletes provider versions.
type ProviderVersionDeleter interface {
	Delete(ctx context.Context, orgID, providerID, id string) error
}

// ProviderVersionCRUD combines all provider version operations.
type ProviderVersionCRUD interface {
	ProviderVersionLister
	ProviderVersionGetter
	ProviderVersionCreator
	ProviderVersionUpdater
	ProviderVersionDeleter
}

// ImplementationLister lists implementations.
type ImplementationLister interface {
	List(ctx context.Context, orgID, providerID, versionID string, opts *ListOptions) ([]*Implementation, error)
}

// ImplementationGetter retrieves a single implementation.
type ImplementationGetter interface {
	Get(ctx context.Context, orgID, providerID, versionID, id string) (*Implementation, error)
}

// ImplementationCreator creates implementations.
type ImplementationCreator interface {
	Create(ctx context.Context, orgID, providerID, versionID string, impl *Implementation) (*Implementation, error)
}

// ImplementationUpdater updates implementations.
type ImplementationUpdater interface {
	Update(ctx context.Context, orgID, providerID, versionID string, impl *Implementation) (*Implementation, error)
}

// ImplementationDeleter deletes implementations.
type ImplementationDeleter interface {
	Delete(ctx context.Context, orgID, providerID, versionID, id string) error
}

// ImplementationCRUD combines all implementation operations.
type ImplementationCRUD interface {
	ImplementationLister
	ImplementationGetter
	ImplementationCreator
	ImplementationUpdater
	ImplementationDeleter
}

// ModuleVersionLister lists module versions.
type ModuleVersionLister interface {
	List(ctx context.Context, orgID, moduleID string, opts *ListOptions) ([]*ModuleVersion, error)
}

// ModuleVersionGetter retrieves a single module version.
type ModuleVersionGetter interface {
	Get(ctx context.Context, orgID, moduleID, id string) (*ModuleVersion, error)
}

// ModuleVersionCreator creates module versions.
type ModuleVersionCreator interface {
	Create(ctx context.Context, orgID, moduleID string, version *ModuleVersion) (*ModuleVersion, error)
}

// ModuleVersionUpdater updates module versions.
type ModuleVersionUpdater interface {
	Update(ctx context.Context, orgID, moduleID string, version *ModuleVersion) (*ModuleVersion, error)
}

// ModuleVersionDeleter deletes module versions.
type ModuleVersionDeleter interface {
	Delete(ctx context.Context, orgID, moduleID, id string) error
}

// ModuleVersionCRUD combines all module version operations.
type ModuleVersionCRUD interface {
	ModuleVersionLister
	ModuleVersionGetter
	ModuleVersionCreator
	ModuleVersionUpdater
	ModuleVersionDeleter
}

// GithubAppTokenLister lists GitHub App tokens.
type GithubAppTokenLister interface {
	List(ctx context.Context, opts *ListOptions) ([]*GithubAppToken, error)
}

// GithubAppTokenGetter retrieves a single GitHub App token.
type GithubAppTokenGetter interface {
	Get(ctx context.Context, id string) (*GithubAppToken, error)
}

// GithubAppTokenCreator creates GitHub App tokens.
type GithubAppTokenCreator interface {
	Create(ctx context.Context, token *GithubAppToken) (*GithubAppToken, error)
}

// GithubAppTokenUpdater updates GitHub App tokens.
type GithubAppTokenUpdater interface {
	Update(ctx context.Context, token *GithubAppToken) (*GithubAppToken, error)
}

// GithubAppTokenDeleter deletes GitHub App tokens.
type GithubAppTokenDeleter interface {
	Delete(ctx context.Context, id string) error
}

// GithubAppTokenCRUD combines all GitHub App token operations.
type GithubAppTokenCRUD interface {
	GithubAppTokenLister
	GithubAppTokenGetter
	GithubAppTokenCreator
	GithubAppTokenUpdater
	GithubAppTokenDeleter
}

// AddressLister lists addresses.
type AddressLister interface {
	List(ctx context.Context, orgID, jobID string, opts *ListOptions) ([]*Address, error)
}

// AddressGetter retrieves a single address.
type AddressGetter interface {
	Get(ctx context.Context, orgID, jobID, id string) (*Address, error)
}

// AddressCreator creates addresses.
type AddressCreator interface {
	Create(ctx context.Context, orgID, jobID string, address *Address) (*Address, error)
}

// AddressUpdater updates addresses.
type AddressUpdater interface {
	Update(ctx context.Context, orgID, jobID string, address *Address) (*Address, error)
}

// AddressDeleter deletes addresses.
type AddressDeleter interface {
	Delete(ctx context.Context, orgID, jobID, id string) error
}

// AddressCRUD combines all address operations.
type AddressCRUD interface {
	AddressLister
	AddressGetter
	AddressCreator
	AddressUpdater
	AddressDeleter
}

// ProjectLister lists projects.
type ProjectLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*Project, error)
}

// ProjectGetter retrieves a single project.
type ProjectGetter interface {
	Get(ctx context.Context, orgID, id string) (*Project, error)
}

// ProjectCreator creates projects.
type ProjectCreator interface {
	Create(ctx context.Context, orgID string, project *Project) (*Project, error)
}

// ProjectUpdater updates projects.
type ProjectUpdater interface {
	Update(ctx context.Context, orgID string, project *Project) (*Project, error)
}

// ProjectDeleter deletes projects.
type ProjectDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// ProjectCRUD combines all project operations.
type ProjectCRUD interface {
	ProjectLister
	ProjectGetter
	ProjectCreator
	ProjectUpdater
	ProjectDeleter
}

// ProjectAccessLister lists project access rules.
type ProjectAccessLister interface {
	List(ctx context.Context, orgID, projectID string, opts *ListOptions) ([]*ProjectAccess, error)
}

// ProjectAccessGetter retrieves a single project access rule.
type ProjectAccessGetter interface {
	Get(ctx context.Context, orgID, projectID, id string) (*ProjectAccess, error)
}

// ProjectAccessCreator creates project access rules.
type ProjectAccessCreator interface {
	Create(ctx context.Context, orgID, projectID string, access *ProjectAccess) (*ProjectAccess, error)
}

// ProjectAccessUpdater updates project access rules.
type ProjectAccessUpdater interface {
	Update(ctx context.Context, orgID, projectID string, access *ProjectAccess) (*ProjectAccess, error)
}

// ProjectAccessDeleter deletes project access rules.
type ProjectAccessDeleter interface {
	Delete(ctx context.Context, orgID, projectID, id string) error
}

// ProjectAccessCRUD combines all project access operations.
type ProjectAccessCRUD interface {
	ProjectAccessLister
	ProjectAccessGetter
	ProjectAccessCreator
	ProjectAccessUpdater
	ProjectAccessDeleter
}

// FederatedLister lists federated identity configurations.
type FederatedLister interface {
	List(ctx context.Context, opts *ListOptions) ([]*Federated, error)
}

// FederatedGetter retrieves a single federated identity configuration.
type FederatedGetter interface {
	Get(ctx context.Context, id string) (*Federated, error)
}

// FederatedCreator creates federated identity configurations.
type FederatedCreator interface {
	Create(ctx context.Context, fed *Federated) (*Federated, error)
}

// FederatedUpdater updates federated identity configurations.
type FederatedUpdater interface {
	Update(ctx context.Context, fed *Federated) (*Federated, error)
}

// FederatedDeleter deletes federated identity configurations.
type FederatedDeleter interface {
	Delete(ctx context.Context, id string) error
}

// FederatedCRUD combines all federated identity operations.
type FederatedCRUD interface {
	FederatedLister
	FederatedGetter
	FederatedCreator
	FederatedUpdater
	FederatedDeleter
}

// FederatedClaimLister lists federated claims.
type FederatedClaimLister interface {
	List(ctx context.Context, federatedID string, opts *ListOptions) ([]*FederatedClaim, error)
}

// FederatedClaimGetter retrieves a single federated claim.
type FederatedClaimGetter interface {
	Get(ctx context.Context, federatedID, id string) (*FederatedClaim, error)
}

// FederatedClaimCreator creates federated claims.
type FederatedClaimCreator interface {
	Create(ctx context.Context, federatedID string, claim *FederatedClaim) (*FederatedClaim, error)
}

// FederatedClaimUpdater updates federated claims.
type FederatedClaimUpdater interface {
	Update(ctx context.Context, federatedID string, claim *FederatedClaim) (*FederatedClaim, error)
}

// FederatedClaimDeleter deletes federated claims.
type FederatedClaimDeleter interface {
	Delete(ctx context.Context, federatedID, id string) error
}

// FederatedClaimCRUD combines all federated claim operations.
type FederatedClaimCRUD interface {
	FederatedClaimLister
	FederatedClaimGetter
	FederatedClaimCreator
	FederatedClaimUpdater
	FederatedClaimDeleter
}

// OperationsSubmitter submits atomic operations.
type OperationsSubmitter interface {
	Submit(ctx context.Context, ops *AtomicRequest) (*AtomicResponse, error)
}

// NotificationConfigurationLister lists organization-level notification configurations.
type NotificationConfigurationLister interface {
	List(ctx context.Context, orgID string, opts *ListOptions) ([]*NotificationConfiguration, error)
}

// NotificationConfigurationWorkspaceLister lists workspace-level notification configurations.
type NotificationConfigurationWorkspaceLister interface {
	ListByWorkspace(ctx context.Context, orgID, workspaceID string, opts *ListOptions) ([]*NotificationConfiguration, error)
}

// NotificationConfigurationGetter retrieves a single notification configuration.
type NotificationConfigurationGetter interface {
	Get(ctx context.Context, orgID, id string) (*NotificationConfiguration, error)
}

// NotificationConfigurationCreator creates organization-level notification configurations.
type NotificationConfigurationCreator interface {
	Create(ctx context.Context, orgID string, config *NotificationConfiguration) (*NotificationConfiguration, error)
}

// NotificationConfigurationWorkspaceCreator creates workspace-level notification configurations.
type NotificationConfigurationWorkspaceCreator interface {
	CreateForWorkspace(ctx context.Context, orgID, workspaceID string, config *NotificationConfiguration) (*NotificationConfiguration, error)
}

// NotificationConfigurationUpdater updates notification configurations.
type NotificationConfigurationUpdater interface {
	Update(ctx context.Context, orgID string, config *NotificationConfiguration) (*NotificationConfiguration, error)
}

// NotificationConfigurationDeleter deletes notification configurations.
type NotificationConfigurationDeleter interface {
	Delete(ctx context.Context, orgID, id string) error
}

// NotificationConfigurationCRUD combines all notification configuration operations.
type NotificationConfigurationCRUD interface {
	NotificationConfigurationLister
	NotificationConfigurationWorkspaceLister
	NotificationConfigurationGetter
	NotificationConfigurationCreator
	NotificationConfigurationWorkspaceCreator
	NotificationConfigurationUpdater
	NotificationConfigurationDeleter
}

// NotificationTriggerLister lists notification triggers.
type NotificationTriggerLister interface {
	List(ctx context.Context, orgID, configID string, opts *ListOptions) ([]*NotificationTrigger, error)
}

// NotificationTriggerGetter retrieves a single notification trigger.
type NotificationTriggerGetter interface {
	Get(ctx context.Context, orgID, configID, id string) (*NotificationTrigger, error)
}

// NotificationTriggerCreator creates notification triggers.
type NotificationTriggerCreator interface {
	Create(ctx context.Context, orgID, configID string, trigger *NotificationTrigger) (*NotificationTrigger, error)
}

// NotificationTriggerUpdater updates notification triggers.
type NotificationTriggerUpdater interface {
	Update(ctx context.Context, orgID, configID string, trigger *NotificationTrigger) (*NotificationTrigger, error)
}

// NotificationTriggerDeleter deletes notification triggers.
type NotificationTriggerDeleter interface {
	Delete(ctx context.Context, orgID, configID, id string) error
}

// NotificationTriggerCRUD combines all notification trigger operations.
type NotificationTriggerCRUD interface {
	NotificationTriggerLister
	NotificationTriggerGetter
	NotificationTriggerCreator
	NotificationTriggerUpdater
	NotificationTriggerDeleter
}

// Compile-time interface satisfaction checks.
var (
	_ OrganizationCRUD              = (*OrganizationService)(nil)
	_ WorkspaceCRUD                 = (*WorkspaceService)(nil)
	_ ModuleCRUD                    = (*ModuleService)(nil)
	_ TeamCRUD                      = (*TeamService)(nil)
	_ TeamTokenCRUD                 = (*TeamTokenService)(nil)
	_ VariableCRUD                  = (*VariableService)(nil)
	_ OrganizationVariableCRUD      = (*OrganizationVariableService)(nil)
	_ TemplateCRUD                  = (*TemplateService)(nil)
	_ TagCRUD                       = (*TagService)(nil)
	_ VCSCRUD                       = (*VCSService)(nil)
	_ SSHCRUD                       = (*SSHService)(nil)
	_ AgentCRUD                     = (*AgentService)(nil)
	_ CollectionCRUD                = (*CollectionService)(nil)
	_ CollectionItemCRUD            = (*CollectionItemService)(nil)
	_ CollectionReferenceCRUD       = (*CollectionReferenceService)(nil)
	_ WorkspaceTagCRUD              = (*WorkspaceTagService)(nil)
	_ WorkspaceAccessCRUD           = (*WorkspaceAccessService)(nil)
	_ WorkspaceScheduleCRUD         = (*WorkspaceScheduleService)(nil)
	_ WebhookCRUD                   = (*WebhookService)(nil)
	_ WebhookEventCRUD              = (*WebhookEventService)(nil)
	_ HistoryCRUD                   = (*HistoryService)(nil)
	_ JobCRUD                       = (*JobService)(nil)
	_ ActionCRUD                    = (*ActionService)(nil)
	_ StepCRUD                      = (*StepService)(nil)
	_ ProviderCRUD                  = (*ProviderService)(nil)
	_ ProviderVersionCRUD           = (*ProviderVersionService)(nil)
	_ ImplementationCRUD            = (*ImplementationService)(nil)
	_ ModuleVersionCRUD             = (*ModuleVersionService)(nil)
	_ GithubAppTokenCRUD            = (*GithubAppTokenService)(nil)
	_ AddressCRUD                   = (*AddressService)(nil)
	_ ProjectCRUD                   = (*ProjectService)(nil)
	_ ProjectAccessCRUD             = (*ProjectAccessService)(nil)
	_ FederatedCRUD                 = (*FederatedService)(nil)
	_ FederatedClaimCRUD            = (*FederatedClaimService)(nil)
	_ OperationsSubmitter           = (*OperationsService)(nil)
	_ NotificationConfigurationCRUD = (*NotificationConfigurationService)(nil)
	_ NotificationTriggerCRUD       = (*NotificationTriggerService)(nil)
)
