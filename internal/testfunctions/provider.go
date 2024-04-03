package testfunctions

import (
	"github.com/apparentlymart/go-tf-func-provider/tffunc"

	"github.com/opentofu/terraform-provider-testfunctions/internal/testfunctions/functions"
)

func NewProvider() *tffunc.Provider {
	p := tffunc.Provider{}

	p.AddFunction("error", functions.ErrorFunc)
	return &p
}
