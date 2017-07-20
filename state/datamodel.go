package state

// "secure secret": marshals all fields to JSON from InsecureSecret *except*
// sensitive fields
type Secret struct {
	Id                     string
	FolderId               string
	Title                  string
	Username               string
	password               string
	sshPrivateKey          string
	SshPublicKeyAuthorized string
	otpProvisioningUrl     string
	Description            string
	// created
	// password last changed
}

type InsecureSecret struct {
	Id                     string
	FolderId               string
	Title                  string
	Username               string
	Password               string
	SshPrivateKey          string
	SshPublicKeyAuthorized string
	OtpProvisioningUrl     string
	Description            string
	// created
	// password last changed
}

func (i *InsecureSecret) ToSecureSecret() Secret {
	return Secret{
		Id:                     i.Id,
		FolderId:               i.FolderId,
		Title:                  i.Title,
		Username:               i.Username,
		password:               i.Password,
		sshPrivateKey:          i.SshPrivateKey,
		SshPublicKeyAuthorized: i.SshPublicKeyAuthorized,
		otpProvisioningUrl:     i.OtpProvisioningUrl,
		Description:            i.Description,
	}
}

func (s *Secret) ToInsecureSecret() InsecureSecret {
	return InsecureSecret{
		Id:                     s.Id,
		FolderId:               s.FolderId,
		Title:                  s.Title,
		Username:               s.Username,
		Password:               s.password,
		SshPrivateKey:          s.sshPrivateKey,
		SshPublicKeyAuthorized: s.SshPublicKeyAuthorized,
		OtpProvisioningUrl:     s.otpProvisioningUrl,
		Description:            s.Description,
	}
}

type Folder struct {
	Id       string
	ParentId string
	Name     string
}

type State struct {
	Password string
	State    *Statefile
}

type Statefile struct {
	Secrets []InsecureSecret
	Folders []Folder
}
