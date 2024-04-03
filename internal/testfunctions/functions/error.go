package functions

import (
	"fmt"

	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
)

var ErrorFunc = &function.Spec{
	Description: "This function always returns an error.",
	Params:      []function.Parameter{},
	Type:        function.StaticReturnType(cty.NilType),
	Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
		return cty.NilVal, fmt.Errorf("this is an error message")
	},
	RefineResult: func(builder *cty.RefinementBuilder) *cty.RefinementBuilder {
		return builder.Null()
	},
}
