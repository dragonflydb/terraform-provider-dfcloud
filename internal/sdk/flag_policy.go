package sdk

// FlagPolicy is a customer-scoped set of Dragonfly flags that can be applied
// to a datastore.
type FlagPolicy struct {
	ID          string   `json:"flag_policy_id"`
	Name        string   `json:"name"`
	Description *string  `json:"description,omitempty"`
	Flags       []string `json:"flags"`
	CreatedAt   int64    `json:"created_at"`
}
