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

	ActionDeploymentList       = "deployment:list"
	ActionDeploymentJsonRead   = "deployment:json:read"
	ActionDeploymentJsonUpdate = "deployment:json:update"
	ActionDeploymentDelete     = "deployment:delete"

	ActionServiceList     = "service:list"
	ActionServiceJsonRead = "service:json:read"
	ActionServiceDelete   = "service:delete"

	ActionConfigMapList       = "configmap:list"
	ActionConfigMapJsonRead   = "configmap:json:read"
	ActionConfigMapJsonUpdate = "configmap:json:update"
	ActionConfigMapDelete     = "configmap:delete"

	ActionSecretList       = "secret:list"
	ActionSecretJsonRead   = "secret:json:read"
	ActionSecretJsonUpdate = "secret:json:update"
	ActionSecretDelete     = "secret:delete"

	ActionNodeList     = "node:list"
	ActionNodeJsonRead = "node:json:read"
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

	ActionDeploymentList: RolesAllUsers,
	ActionServiceList:    RolesAllUsers,
	ActionConfigMapList:  RolesAllUsers,
	ActionSecretList:     RolesAllUsers,
	ActionNodeList:       RolesAllUsers,

	ActionPodLogsRead:           RolesDevPlus,
	ActionContainerJobsList:     RolesDevPlus,
	ActionContainerFilesList:    RolesDevPlus,
	ActionContainerFileRead:     RolesDevPlus,
	ActionContainerFileDownload: RolesDevPlus,

	ActionDeploymentJsonRead: RolesDevPlus,
	ActionServiceJsonRead:    RolesDevPlus,
	ActionConfigMapJsonRead:  RolesDevPlus,
	ActionSecretJsonRead:     RolesDevPlus,
	ActionNodeJsonRead:       RolesDevPlus,

	ActionDeploymentJsonUpdate: RolesDevPlus,
	ActionDeploymentDelete:     RolesDevPlus,
	ActionServiceDelete:        RolesDevPlus,
	ActionConfigMapJsonUpdate:  RolesDevPlus,
	ActionConfigMapDelete:      RolesDevPlus,
	ActionSecretJsonUpdate:     RolesDevPlus,
	ActionSecretDelete:         RolesDevPlus,

	ActionAutomationRun:      RolesOpsPlus,
	ActionAutomationWebhook:  append(RolesOpsPlus, RoleSystem, RoleWebhook),
	ActionContainerJobRun:    RolesDevPlus,
	ActionContainerFileWrite: RolesDevPlus,
}
