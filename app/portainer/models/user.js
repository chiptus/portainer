export function UserViewModel(data) {
  this.Id = data.Id;
  this.Username = data.Username;
  this.Role = data.Role;
  this.EndpointAuthorizations = data.EndpointAuthorizations;
  this.PortainerAuthorizations = data.PortainerAuthorizations;
  this.ThemeSettings = data.ThemeSettings;
  if (data.Role === 1) {
    this.RoleName = 'administrator';
  } else {
    this.RoleName = 'user';
  }
  this.AuthenticationMethod = data.AuthenticationMethod;
  this.Checked = false;
}

export function UserTokenModel(data) {
  this.id = data.id;
  this.userId = data.userId;
  this.description = data.description;
  this.prefix = data.prefix;
  this.dateCreated = data.dateCreated;
  this.lastUsed = data.lastUsed;
}
