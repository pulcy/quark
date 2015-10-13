package vultr

func (vp *vultrProvider) ShowDomainRecords(domain string) error {
	return maskAny(NotImplementedError)
}
