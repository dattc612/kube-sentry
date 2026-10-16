package router

const (
	ActionNamespaceList         = "namespace:list"
	ActionAutomationList        = "automation:list"
	ActionAutomationRead        = "automation:read"
	ActionAutomationRun         = "automation:run"
	ActionAutomationWebhook     = "automation:webhook:trigger"
	ActionExecutionList         = "execution:list"
	ActionExecutionStepsRead    = "execution:steps:read"
	ActionExecutionRead         = "execution:read"
	ActionExecutionStatusRead   = "execution:status:read"
	ActionExecutionLogsRead     = "execution:logs:read"
	ActionPodList               = "pod:list"
	ActionPodLogsRead           = "pod:logs:read"
	ActionContainerJobsList     = "container:jobs:list"
	ActionContainerJobRun       = "container:job:run"
	ActionContainerFilesList    = "container:files:list"
	ActionContainerFileWrite    = "container:file:write"
	ActionContainerFileRead     = "container:file:read"
	ActionContainerFileDownload = "container:file:download"
)

const (
	RoleAdmin     = "admin"
	RoleOperator  = "operator"
	RoleDeveloper = "developer"
	RoleViewer    = "viewer"
	RoleSystem    = "system"
	RoleWebhook   = "webhook"
)

var (
	RolesAllUsers = []string{RoleAdmin, RoleOperator, RoleDeveloper, RoleViewer}
	RolesDevPlus  = []string{RoleAdmin, RoleOperator, RoleDeveloper}
	RolesOpsPlus  = []string{RoleAdmin, RoleOperator}
)

var ActionRolePermissions = map[string][]string{
	ActionNamespaceList:       RolesAllUsers,
	ActionAutomationList:      RolesAllUsers,
	ActionAutomationRead:      RolesAllUsers,
	ActionExecutionList:       RolesAllUsers,
	ActionExecutionRead:       RolesAllUsers,
	ActionExecutionStepsRead:  RolesAllUsers,
	ActionExecutionStatusRead: RolesAllUsers,
	ActionExecutionLogsRead:   RolesAllUsers,
	ActionPodList:             RolesAllUsers,

	ActionPodLogsRead:           RolesDevPlus,
	ActionContainerJobsList:     RolesDevPlus,
	ActionContainerFilesList:    RolesDevPlus,
	ActionContainerFileRead:     RolesDevPlus,
	ActionContainerFileDownload: RolesDevPlus,

	ActionAutomationRun:      RolesOpsPlus,
	ActionAutomationWebhook:  append(RolesOpsPlus, RoleSystem, RoleWebhook),
	ActionContainerJobRun:    RolesDevPlus,
	ActionContainerFileWrite: RolesDevPlus,
}
