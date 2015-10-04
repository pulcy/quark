package vagrant

import (
	"github.com/juju/errgo"
)

var (
	NotImplementedError = errgo.New("not implemented")
	maskAny             = errgo.MaskFunc(errgo.Any)
)
