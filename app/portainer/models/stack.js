import { ResourceControlViewModel } from '@/react/portainer/access-control/models/ResourceControlViewModel';

export function StackViewModel(data) {
  this.Id = data.Id;
  this.Type = data.Type;
  this.Name = data.Name;
  this.EndpointId = data.EndpointId;
  this.SwarmId = data.SwarmId;
  this.Env = data.Env ? data.Env : [];
  this.Option = data.Option;
  this.IsComposeFormat = data.IsComposeFormat;
  if (data.ResourceControl && data.ResourceControl.Id !== 0) {
    this.ResourceControl = new ResourceControlViewModel(data.ResourceControl);
  }
  this.Status = data.Status;
  this.CreationDate = data.CreationDate;
  this.CreatedBy = data.CreatedBy;
  this.UpdateDate = data.UpdateDate;
  this.UpdatedBy = data.UpdatedBy;

  this.Regular = true;
  this.External = false;
  this.Orphaned = false;
  this.Checked = false;
  this.GitConfig = data.GitConfig;
  this.FromAppTemplate = data.FromAppTemplate;
  this.AdditionalFiles = data.AdditionalFiles;
  this.AutoUpdate = data.AutoUpdate;
  this.Webhook = data.Webhook;
  this.StackFileVersion = data.StackFileVersion;
  this.PreviousDeploymentInfo = data.PreviousDeploymentInfo;
}

export function ExternalStackViewModel(name, type, creationDate) {
  this.Name = name;
  this.Type = type;
  this.CreationDate = creationDate;

  this.Regular = false;
  this.External = true;
  this.Orphaned = false;
  this.Checked = false;
}

export function OrphanedStackViewModel(data) {
  this.Id = data.Id;
  this.Type = data.Type;
  this.Name = data.Name;
  this.EndpointId = data.EndpointId;
  this.SwarmId = data.SwarmId;
  this.Env = data.Env ? data.Env : [];
  this.Option = data.Option;
  if (data.ResourceControl && data.ResourceControl.Id !== 0) {
    this.ResourceControl = new ResourceControlViewModel(data.ResourceControl);
  }
  this.Status = data.Status;
  this.CreationDate = data.CreationDate;
  this.CreatedBy = data.CreatedBy;
  this.UpdateDate = data.UpdateDate;
  this.UpdatedBy = data.UpdatedBy;

  this.Regular = false;
  this.External = false;
  this.Orphaned = true;
  this.OrphanedRunning = false;
  this.Checked = false;
}
