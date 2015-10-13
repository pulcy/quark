package vultr

func (vp *vultrProvider) CreateAnsibleHosts(domain string, sshPort int, developersJSON string) error {
	return maskAny(NotImplementedError)
}
