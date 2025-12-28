package resources

type PackageManager interface {
	IsInstalled(name string) (bool, error)
	Install(name string) error
	Remove(name string) error
}

type PackageResource struct {
	CanonicalID string
	PackageName string
	State       string
	Provider    PackageManager
}

func (p *PackageResource) ID() string {
	return p.CanonicalID
}

func (p *PackageResource) Check() (bool, error) {
	return p.Provider.IsInstalled(p.PackageName)
}

func (p *PackageResource) Apply() error {
	if p.State == "absent" {
		return p.Provider.Remove(p.PackageName)
	}
	return p.Provider.Install(p.PackageName)
}
