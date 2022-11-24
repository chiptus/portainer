package packages

import (
	"fmt"
)

func (ghPackages *Packages) DeleteManifest(packageName, digest string) error {
	gpVersions, err := ghPackages.GetPackageVersions(packageName)
	if err != nil {
		return err
	}

	for _, gpVersion := range gpVersions {
		if gpVersion.Name == digest {
			if len(gpVersions) == 1 {
				err = ghPackages.DeletePackage(packageName)
				return err
			} else {
				err = ghPackages.DeletePackageVersion(packageName, gpVersion.ID)
				return err
			}
		}
	}

	return fmt.Errorf("package %s with digest %s not found", packageName, digest)
}
