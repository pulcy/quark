package vultr

import (
	"github.com/juju/errgo"
)

var (
	NotImplementedError  = errgo.New("not implemented")
	InvalidArgumentError = errgo.New("invalid argument")
	maskAny              = errgo.MaskFunc(errgo.Any)
)
