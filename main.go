package main

import (
	"context"
	"fmt"
	"os"
	"strconv"

	"github.com/Shopify/go-lua"
	"github.com/apparentlymart/go-tf-func-provider/tffunc"
	"github.com/zclconf/go-cty/cty"
	"github.com/zclconf/go-cty/cty/function"
	"github.com/zclconf/go-cty/cty/gocty"
)

func main() {
	ctx := context.Background()
	provider := tffunc.NewProvider()
	provider.AddFunction("lua", &function.Spec{
		Description: "This function executes a lua main function with the given parameters",
		Params: []function.Parameter{{
			Name:        "code",
			Description: "Lua Code",
			Type:        cty.String,
		}, {
			Name:        "function",
			Description: "Lua function name",
			Type:        cty.String,
		}},
		VarParam: &function.Parameter{
			Name:        "parameters",
			Description: "Variable parameters passed into the main function",
			Type:        cty.DynamicPseudoType,
		},
		Type: function.StaticReturnType(cty.DynamicPseudoType),
		Impl: func(args []cty.Value, retType cty.Type) (cty.Value, error) {
			code := args[0].AsString()
			function := args[1].AsString()
			args = args[2:]

			l := lua.NewState()
			lua.OpenLibraries(l)

			// Define main function
			if err := lua.DoString(l, code); err != nil {
				return cty.NilVal, err
			}

			// Setup main call
			l.Global(function)
			//if !l.IsFunction(1) {
			//	panic(l.TypeOf(1))
			//}

			for _, arg := range args {
				err := CtyToLua(arg, l)
				if err != nil {
					return cty.NilVal, err
				}
			}

			// Call main, expecting one return value
			l.Call(len(args), 1)

			// Retrieve result
			return LuaToCty(l)
		},
	})
	err := provider.Serve(ctx)
	if err != nil {
		os.Stderr.WriteString(fmt.Sprintf("Error starting provider: %s\n", err))
		os.Exit(1)
	}
}

func CtyToLua(arg cty.Value, l *lua.State) error {
	switch t := arg.Type(); t {
	case cty.Number:
		var v float64
		err := gocty.FromCtyValue(arg, &v)
		if err != nil {
			return err
		}
		l.PushNumber(v)
		return nil
	case cty.String:
		var v string
		err := gocty.FromCtyValue(arg, &v)
		if err != nil {
			return err
		}
		l.PushString(v)
		return nil
	case cty.Bool:
		var v bool
		err := gocty.FromCtyValue(arg, &v)
		if err != nil {
			return err
		}
		l.PushBoolean(v)
		return nil
	default:
		if t.IsObjectType() || t.IsMapType() {
			l.NewTable()
			for k, v := range arg.AsValueMap() {
				l.PushString(k)
				err := CtyToLua(v, l)
				if err != nil {
					return err
				}
				l.SetTable(-3)
			}
			return nil
		}
		if t.IsListType() || t.IsSetType() || t.IsTupleType() {
			l.NewTable()
			for k, v := range arg.AsValueSlice() {
				l.PushInteger(k)
				err := CtyToLua(v, l)
				if err != nil {
					return err
				}
				l.SetTable(-3)
			}
			return nil
		}
		return fmt.Errorf("unsupported parameter type %#v", arg.Type())
	}
}

func LuaToCty(l *lua.State) (cty.Value, error) {
	if l.IsNone(-1) {
		return cty.NilVal, fmt.Errorf("none value should not be returned")
	}

	switch t := l.TypeOf(-1); t {
	case lua.TypeNil:
		return cty.NilVal, nil
	case lua.TypeBoolean:
		return cty.BoolVal(l.ToBoolean(-1)), nil
	case lua.TypeNumber:
		number, _ := l.ToNumber(-1)
		return cty.NumberFloatVal(number), nil
	case lua.TypeString:
		str, _ := l.ToString(-1)
		return cty.StringVal(str), nil
	case lua.TypeTable:
		// https://stackoverflow.com/a/6142700
		mv := make(map[string]cty.Value)

		// Space for key
		l.PushNil()

		// Push value
		for l.Next(-2) {
			// Copy key to top of stack
			l.PushValue(-2)

			// Decode key (also modifies)
			key, ok := l.ToString(-1)
			if !ok {
				return cty.NilVal, fmt.Errorf("bad table index")
			}

			l.Pop(1)

			// Decode Value (also modifies)
			val, err := LuaToCty(l)
			if err != nil {
				return cty.NilVal, err
			}
			mv[key] = val

			l.Pop(1)
		}

		av := make([]cty.Value, len(mv))

		off := 1
		// Hack in an off-by-one offset
		if _, ok := mv["0"]; ok {
			off = 0
		}

		// This is inefficient, but it works
		for i := off; i < len(av)+off; i++ {
			if v, ok := mv[strconv.Itoa(i)]; ok {
				av[i] = v
			} else {
				// Not a coherent list
				return cty.ObjectVal(mv), nil
			}
		}
		return cty.ListVal(av), nil
	default:
		return cty.NilVal, fmt.Errorf("unhanded return type %s!", t)
	}
}
