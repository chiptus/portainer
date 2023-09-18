package datastore

import (
	"github.com/portainer/portainer-ee/api/dataservices"
	portainer "github.com/portainer/portainer/api"
)

type StoreTx struct {
	store *Store
	tx    portainer.Transaction
}

func (tx *StoreTx) IsErrObjectNotFound(err error) bool {
	return tx.store.IsErrObjectNotFound(err)
}

func (tx *StoreTx) CustomTemplate() dataservices.CustomTemplateService { return nil }

func (tx *StoreTx) PendingActions() dataservices.PendingActionsService { return nil }

func (tx *StoreTx) EdgeAsyncCommand() dataservices.EdgeAsyncCommandService {
	return tx.store.EdgeAsyncCommandService.Tx(tx.tx)
}

func (tx *StoreTx) EdgeConfig() dataservices.EdgeConfigService {
	return tx.store.EdgeConfigService.Tx(tx.tx)
}

func (tx *StoreTx) EdgeConfigState() dataservices.EdgeConfigStateService {
	return tx.store.EdgeConfigStateService.Tx(tx.tx)
}

func (tx *StoreTx) CloudProvisioning() dataservices.CloudProvisioningService { return nil }

func (tx *StoreTx) EdgeGroup() dataservices.EdgeGroupService {
	return tx.store.EdgeGroupService.Tx(tx.tx)
}

func (tx *StoreTx) EdgeJob() dataservices.EdgeJobService {
	return tx.store.EdgeJobService.Tx(tx.tx)
}

func (tx *StoreTx) EdgeUpdateSchedule() dataservices.EdgeUpdateScheduleService {
	return tx.store.EdgeUpdateScheduleService.Tx(tx.tx)
}

func (tx *StoreTx) EdgeStack() dataservices.EdgeStackService {
	return tx.store.EdgeStackService.Tx(tx.tx)
}

func (tx *StoreTx) EdgeStackLog() dataservices.EdgeStackLogService {
	return tx.store.EdgeStackLogService.Tx(tx.tx)
}

func (tx *StoreTx) Endpoint() dataservices.EndpointService {
	return tx.store.EndpointService.Tx(tx.tx)
}

func (tx *StoreTx) EndpointGroup() dataservices.EndpointGroupService {
	return tx.store.EndpointGroupService.Tx(tx.tx)
}

func (tx *StoreTx) EndpointRelation() dataservices.EndpointRelationService {
	return tx.store.EndpointRelationService.Tx(tx.tx)
}

func (tx *StoreTx) Enforcement() dataservices.EnforcementService               { return nil }
func (tx *StoreTx) FDOProfile() dataservices.FDOProfileService                 { return nil }
func (tx *StoreTx) HelmUserRepository() dataservices.HelmUserRepositoryService { return nil }
func (tx *StoreTx) License() dataservices.LicenseRepository                    { return nil }

func (tx *StoreTx) Registry() dataservices.RegistryService {
	return tx.store.RegistryService.Tx(tx.tx)
}

func (tx *StoreTx) ResourceControl() dataservices.ResourceControlService {
	return tx.store.ResourceControlService.Tx(tx.tx)
}

func (tx *StoreTx) Role() dataservices.RoleService {
	return tx.store.RoleService.Tx(tx.tx)
}

func (tx *StoreTx) APIKeyRepository() dataservices.APIKeyRepository { return nil }

func (tx *StoreTx) GitCredential() dataservices.GitCredential {
	return tx.store.GitCredentialService.Tx(tx.tx)
}

func (tx *StoreTx) S3Backup() dataservices.S3BackupService { return nil }

func (tx *StoreTx) Settings() dataservices.SettingsService {
	return tx.store.SettingsService.Tx(tx.tx)
}

func (tx *StoreTx) Snapshot() dataservices.SnapshotService {
	return tx.store.SnapshotService.Tx(tx.tx)
}

func (tx *StoreTx) SSLSettings() dataservices.SSLSettingsService { return nil }

func (tx *StoreTx) Stack() dataservices.StackService {
	return tx.store.StackService.Tx(tx.tx)
}

func (tx *StoreTx) Tag() dataservices.TagService {
	return tx.store.TagService.Tx(tx.tx)
}

func (tx *StoreTx) TeamMembership() dataservices.TeamMembershipService {
	return tx.store.TeamMembershipService.Tx(tx.tx)
}

func (tx *StoreTx) Team() dataservices.TeamService                 { return nil }
func (tx *StoreTx) PodSecurity() dataservices.PodSecurityService   { return nil }
func (tx *StoreTx) TunnelServer() dataservices.TunnelServerService { return nil }

func (tx *StoreTx) User() dataservices.UserService {
	return tx.store.UserService.Tx(tx.tx)
}

func (tx *StoreTx) Version() dataservices.VersionService { return nil }
func (tx *StoreTx) Webhook() dataservices.WebhookService { return nil }

func (tx *StoreTx) CloudCredential() dataservices.CloudCredentialService {
	return tx.store.CloudCredentialService.Tx(tx.tx)
}
