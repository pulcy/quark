package cloudflare

import (
	"github.com/juju/errgo"
)

var (
	DomainNotFoundError = errgo.New("domain not found")
	maskAny             = errgo.MaskFunc(errgo.Any)
)
