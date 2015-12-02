package digitalocean

import (
	"github.com/juju/errgo"
)

var (
	NotFoundError       = errgo.New("not found")
	NotImplementedError = errgo.New("not implemented")
	maskAny             = errgo.MaskFunc(errgo.Any)
)
