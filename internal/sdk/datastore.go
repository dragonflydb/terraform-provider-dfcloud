package sdk

type CloudProvider string

const (
	CloudProviderAWS   CloudProvider = "aws"
	CloudProviderGCP   CloudProvider = "gcp"
	CloudProviderAzure CloudProvider = "azure"
)

// DatastoreLocation represents where the datastore should be provisioned.
type DatastoreLocation struct {
	Provider CloudProvider `json:"provider"`
	Region   string        `json:"region"`
	// AvailabilityZones indicates which availability zones the datastore
	// should use in priority order.
	AvailabilityZones []string `json:"availability_zones" mapstructure:"availability_zones"`
}

type PerformanceTier string
type AclRuleArray []string

const (
	PerformanceTierDev      PerformanceTier = "dev"
	PerformanceTierStandard PerformanceTier = "standard"
	PerformanceTierEnhanced PerformanceTier = "enhanced"
)

var PerformanceTiers = []PerformanceTier{
	PerformanceTierDev,
	PerformanceTierStandard,
	PerformanceTierEnhanced,
}

func PerformanceTiersString() []string {
	var ss []string
	for _, tier := range PerformanceTiers {
		ss = append(ss, string(tier))
	}
	return ss
}

type DatastoreTier struct {
	// Memory is the maximum number of bytes Dragonfly can consume.
	Memory uint64 `json:"max_memory_bytes"`
	// PerformanceTier determines number of CPUs provisioned relative to
	// memory.
	PerformanceTier PerformanceTier `json:"performance_tier"`

	// Replicas is the number of Dragonfly replicas (not including the master).
	Replicas *int `json:"replicas"`
}

type DatastoreDragonflyConfig struct {
	CacheMode *bool         `json:"cache_mode"`
	TLS       *bool         `json:"tls"`
	BullMQ    *bool         `json:"bullmq"`
	Sidekiq   *bool         `json:"sidekiq"`
	Memcached *bool         `json:"memcached"`
	AclRules  *AclRuleArray `json:"acl_rules"`
}

// DatastoreConfig contains the datastores configurable fields.
type DatastoreConfig struct {
	Name string `json:"name"`
	// NetworkID is an optional ID of a dedicated network to provision the
	// datastore in.
	NetworkID string            `json:"network_id"`
	Location  DatastoreLocation `json:"location"`
	Tier      DatastoreTier     `json:"tier"`
	// Dragonfly contains the Dragonfly node configuration.
	Dragonfly DatastoreDragonflyConfig `json:"dragonfly"`

	BackupPolicy BackupPolicy `json:"backup_policy" mapstructure:"backup_policy"`

	Restore RestoreBackup `json:"restore"`

	DisablePasskey bool `json:"disable_passkey"`
}

type RestoreBackup struct {
	// Backup contains the ID of the backup to restore.
	BackupId string `json:"backup_id"`
	// Loaded denotes if the backup is loaded onto the datastore
	Loaded bool `json:"loaded"`
}

type BackupPolicy struct {
	Enabled   *bool `json:"enabled"`
	Retention int   `json:"retention,omitempty"`
	EveryHour *bool `json:"every_hour,omitempty"`
	EveryDay  *bool `json:"every_day,omitempty"`
	Hours     []int `json:"hours,omitempty"`
	WeekDays  []int `json:"weekdays,omitempty"`
}

type DatastoreDashboard struct {
	// URL contains the datastores public Grafana dashboard URL.
	URL string `json:"url"`
}

// DatastoreStatus represents the current status of the datastore.
type DatastoreStatus string

const (
	// DatastoreStatusPending is set when the user has requested the datastore
	// and it is being asynchronously provisioned.
	DatastoreStatusPending DatastoreStatus = "pending"
	// DatastoreStatusUpdating is set when a user has requested an update and
	// it is being asynchronously provisioned.
	DatastoreStatusUpdating DatastoreStatus = "updating"
	// DatastoreStatusRestoring is set when a user has requested a backup
	// that is being asyncronously restored.
	DatastoreStatusRestoring DatastoreStatus = "restoring"
	// DatastoreStatusActive is set when the datastore has been provisioned and
	// is usable.
	DatastoreStatusActive DatastoreStatus = "active"
	// DatastoreStatusDeleting is set when the user has requested the datastore
	// to be deleted and it is being asynchronously deprovisioned.
	DatastoreStatusDeleting DatastoreStatus = "deleting"
	// DatastoreStatusDeleted is set when the datastore has been deprovisioned.
	DatastoreStatusDeleted DatastoreStatus = "deleted"
)

type Datastore struct {
	// ID is a unique identifier for the datastore.
	ID string `json:"datastore_id"`

	Status DatastoreStatus `json:"status"`

	CreatedAt int64 `json:"created_at" mapstructure:"created_at"`

	// Key is the Dragonfly key to configure when connecting to your
	// datastore.
	Key string `json:"password"`

	// Addr is the hostname and port of your datastore.
	Addr string `json:"addr"`

	// Dashboard contains details on the datastores public Grafana dashboard.
	Dashboard *DatastoreDashboard `json:"dashboard"`

	Config DatastoreConfig `json:"config"`
}
