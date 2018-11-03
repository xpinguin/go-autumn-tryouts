// Opaque prolog atom for holding golang's values
package browsemon

import (
	"fmt"

	"github.com/mndrix/golog/term"
)

type GoValue struct {
	v interface{}
}

func (v GoValue) AsInterface() interface{} {
	return v.v
}

func (v GoValue) String() string {
	return term.QuoteFunctor(fmt.Sprint(v.v))
}
func (v GoValue) Type() int {
	return term.AtomType
}
func (v GoValue) Indicator() string {
	return fmt.Sprintf("%s/0", v.String())
}

func (v GoValue) ReplaceVariables(env term.Bindings) term.Term {
	return v
}

func (v GoValue) Unify(env term.Bindings, x term.Term) (term.Bindings, error) {
	switch other := x.(type) {
	case *term.Variable:
		return other.Unify(env, v)
	case GoValue:
		if v.v == other.v {
			return env, nil
		}
		return env, term.CantUnify
	default:
		return env, term.CantUnify
	}
}
