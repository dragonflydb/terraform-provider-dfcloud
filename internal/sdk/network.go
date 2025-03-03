package sdk

// NetworkStatus represents the current status of the network.
type NetworkStatus string

const (
	// NetworkStatusPending is set when the user has requested the network
	// and it is being asynchronously provisioned.
	NetworkStatusPending NetworkStatus = "pending"
	// DatastoreStatusActive is set when the network has been provisioned.
	NetworkStatusActive NetworkStatus = "active"
	// DatastoreStatusActive is set when the network was requested but could
	// not be provisioned.
	NetworkStatusFailed NetworkStatus = "failed"
	// NetworkStatusDeleting is set when the user has requested the network to
	// be deleted and it is being asynchronously deprovisioned.
	NetworkStatusDeleting NetworkStatus = "deleting"
	// NetworkStatusDeleted is set when the network has been deprovisioned.
	NetworkStatusDeleted NetworkStatus = "deleted"
)

// NetworkLocation represents where the network should be provisioned.
type NetworkLocation struct {
	Provider CloudProvider `json:"provider"`
	Region   string        `json:"region"`
}

type NetworkVPC struct {
	// ResourceID is the ID of the VPC.
	ResourceID string `json:"resource_id"`
	// AccountID is the Dragonfly Cloud account ID that owns the VPC resource.
	// This is required to setup peering connections.
	AccountID string `json:"account_id"`
}

type NetworkConfig struct {
	Name      string          `json:"name"`
	Location  NetworkLocation `json:"location"`
	CIDRBlock string          `json:"cidr_block"`
}

type Network struct {
	ID string `json:"network_id"`

	Status NetworkStatus `json:"status"`

	CreatedAt int64 `json:"created_at"`

	// VPC contains details on the networks provisioned VPC. This is required
	// to setup VPC peering.
	VPC *NetworkVPC `json:"vpc"`

	*NetworkConfig
}
