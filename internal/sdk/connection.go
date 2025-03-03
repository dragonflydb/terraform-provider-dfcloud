package sdk

// NetworkStatus represents the current status of the connection.
type ConnectionStatus string

const (
	// ConnectionStatusPending is set when the user has requested and it is
	// being asynchronously setup.
	ConnectionStatusPending ConnectionStatus = "pending"
	// ConnectionStatusActive indicates the peer connection has been
	// established.
	ConnectionStatusActive ConnectionStatus = "active"
	// ConnectionStatusInactive indicates the peer connection has not yet been
	// approved on the customers account.
	ConnectionStatusInactive ConnectionStatus = "inactive"
	// ConnectionStatusIrrecoverable indicates the peer connection was deleted
	// from the customers account.
	ConnectionStatusIrrecoverable ConnectionStatus = "irrecoverable"
	// ConnectionStatusDeleting is set when the user has requested the
	// connection to be deleted and it is being asynchronously deprovisioned.
	ConnectionStatusDeleting ConnectionStatus = "deleting"
	// ConnectionStatusDeleting is set when the network has been deprovisioned.
	ConnectionStatusDeleted ConnectionStatus = "deleted"
	// ConnectionStatusFailed indicates the peer connection was requested but
	// could not be connected.
	ConnectionStatusFailed ConnectionStatus = "failed"
)

// PeerConfig describes the VPC to connect to.
type PeerConfig struct {
	// AccountID is the account ID of the target VPC.
	AccountID string `json:"account_id"`
	// CIDRBlock is the CIDR block of the target VPC.
	CIDRBlock string `json:"cidr_block"`
	// VPCID is the ID of the target VPC.
	VPCID string `json:"vpc_id"`
	// Region is the region of the target VPC. Only specify if the target VPC
	// is in a different region to the network your connecting to.
	Region string `json:"region,omitempty"`
}

type ConnectionConfig struct {
	Name      string     `json:"name"`
	NetworkID string     `json:"network_id"`
	Peer      PeerConfig `json:"peer"`
}

// Connection represents a network peer-connection.
type Connection struct {
	ID string `json:"connection_id"`

	Status ConnectionStatus `json:"status"`

	// StatusDetail provides more information on the status of the connection.
	StatusDetail string `json:"status_detail,omitempty"`

	// PeerConnectionID is the underlying cloud provider peer connection ID.
	PeerConnectionID string `json:"peer_connection_id"`

	Config *ConnectionConfig `json:"connection_config"`
}
