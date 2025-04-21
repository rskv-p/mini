package rtm_type

import "github.com/rskv-p/mini/mod/m_act/act_type"

// RuntimeClient defines the contract for interacting with external systems from within a runtime.
type RuntimeClient interface {
	ExecuteAction(action act_type.IAction) (any, error) // Execute an action in the runtime.
}
