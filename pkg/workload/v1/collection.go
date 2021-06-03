package v1

func (c WorkloadCollection) GetDomain() string {
	return c.Spec.Domain
}

func (c WorkloadCollection) HasRootCmdName() bool {
	if c.Spec.CompanionCliRootcmd.Name != "" {
		return true
	} else {
		return false
	}
}

func (c WorkloadCollection) GetRootCmdName() string {
	return c.Spec.CompanionCliRootcmd.Name
}

func (c WorkloadCollection) GetRootCmdDescr() string {
	return c.Spec.CompanionCliRootcmd.Description
}
