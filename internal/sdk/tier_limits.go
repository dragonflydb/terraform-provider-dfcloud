package sdk

import "fmt"

// Byte size constants.
const (
	MB uint64 = 1_000_000
	GB uint64 = 1_000_000_000
)

// DevMaxClusterShards is the maximum number of shards allowed for
// a dev-tier cluster datastore.
const DevMaxClusterShards = 10

// allProviders is a shorthand for all supported cloud providers.
var allProviders = []CloudProvider{CloudProviderAWS, CloudProviderGCP, CloudProviderAzure}

// ShardSizeRule maps a single shard size to the cloud providers that
// support it. Use allProviders when every provider is supported.
type ShardSizeRule struct {
	SizeBytes uint64
	Providers []CloudProvider
}

// TierShardSizes defines the supported single-shard (per-node) memory
// sizes for each performance tier, along with which cloud providers
// support that size.
//
// For cluster mode the shard_memory value is validated against this
// table, while max_memory_bytes is validated for divisibility.
var TierShardSizes = map[PerformanceTier][]ShardSizeRule{
	PerformanceTierDev: {
		{SizeBytes: 3 * GB, Providers: allProviders},
	},
	PerformanceTierStandard: {
		{SizeBytes: 12*GB + 500*MB, Providers: allProviders},                                                      // 12.5 GB
		{SizeBytes: 25 * GB, Providers: allProviders},                                                             // 25 GB
		{SizeBytes: 50 * GB, Providers: allProviders},                                                             // 50 GB
		{SizeBytes: 100 * GB, Providers: allProviders},                                                            // 100 GB
		{SizeBytes: 200 * GB, Providers: allProviders},                                                            // 200 GB
		{SizeBytes: 300 * GB, Providers: []CloudProvider{CloudProviderGCP, CloudProviderAzure}},                   // 300 GB
		{SizeBytes: 400 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP, CloudProviderAzure}}, // 400 GB
		{SizeBytes: 500 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},                     // 500 GB
		{SizeBytes: 600 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},                     // 600 GB
	},
	PerformanceTierEnhanced: {
		{SizeBytes: 6*GB + 250*MB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},                // 6.25 GB
		{SizeBytes: 6*GB + 500*MB, Providers: []CloudProvider{CloudProviderAzure}},                                // 6.5 GB
		{SizeBytes: 12*GB + 500*MB, Providers: allProviders},                                                      // 12.5 GB
		{SizeBytes: 25 * GB, Providers: allProviders},                                                             // 25 GB
		{SizeBytes: 50 * GB, Providers: allProviders},                                                             // 50 GB
		{SizeBytes: 100 * GB, Providers: allProviders},                                                            // 100 GB
		{SizeBytes: 150 * GB, Providers: []CloudProvider{CloudProviderGCP, CloudProviderAzure}},                   // 150 GB
		{SizeBytes: 200 * GB, Providers: allProviders},                                                            // 200 GB
		{SizeBytes: 250 * GB, Providers: []CloudProvider{CloudProviderGCP}},                                       // 250 GB
		{SizeBytes: 300 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP, CloudProviderAzure}}, // 300 GB
		{SizeBytes: 400 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},                     // 400 GB
	},
	PerformanceTierExtreme: {
		{SizeBytes: 6*GB + 250*MB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}}, // 6.25 GB
		{SizeBytes: 6*GB + 500*MB, Providers: []CloudProvider{CloudProviderAzure}},                 // 6.5 GB
		{SizeBytes: 12*GB + 500*MB, Providers: allProviders},                                       // 12.5 GB
		{SizeBytes: 25 * GB, Providers: allProviders},                                              // 25 GB
		{SizeBytes: 50 * GB, Providers: allProviders},                                              // 50 GB
		{SizeBytes: 100 * GB, Providers: allProviders},                                             // 100 GB
		{SizeBytes: 150 * GB, Providers: []CloudProvider{CloudProviderGCP}},                        // 150 GB
		{SizeBytes: 200 * GB, Providers: allProviders},                                             // 200 GB
	},
	PerformanceTierBYOC: {
		{SizeBytes: 6*GB + 250*MB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},  // 6.25 GB
		{SizeBytes: 12*GB + 500*MB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}}, // 12.5 GB
		{SizeBytes: 25 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},        // 25 GB
		{SizeBytes: 50 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},        // 50 GB
		{SizeBytes: 100 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},       // 100 GB
		{SizeBytes: 200 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},       // 200 GB
		{SizeBytes: 300 * GB, Providers: []CloudProvider{CloudProviderGCP}},                         // 300 GB
		{SizeBytes: 400 * GB, Providers: []CloudProvider{CloudProviderAWS, CloudProviderGCP}},       // 400 GB
	},
}

// IsSupportedShardSize reports whether the given byte count is a supported
// shard size for the given performance tier and cloud provider.
func IsSupportedShardSize(tier PerformanceTier, sizeBytes uint64, provider CloudProvider) bool {
	rules, ok := TierShardSizes[tier]
	if !ok {
		return false
	}
	for _, rule := range rules {
		if rule.SizeBytes != sizeBytes {
			continue
		}
		for _, p := range rule.Providers {
			if p == provider {
				return true
			}
		}
	}
	return false
}

// SupportedSizesForTier returns the list of supported shard sizes (in bytes)
// for the given tier and cloud provider combination.
func SupportedSizesForTier(tier PerformanceTier, provider CloudProvider) []uint64 {
	rules, ok := TierShardSizes[tier]
	if !ok {
		return nil
	}
	var sizes []uint64
	for _, rule := range rules {
		for _, p := range rule.Providers {
			if p == provider {
				sizes = append(sizes, rule.SizeBytes)
				break
			}
		}
	}
	return sizes
}

// ValidCloudProviders returns a list of supported cloud provider strings.
func ValidCloudProviders() []string {
	return []string{
		string(CloudProviderAWS),
		string(CloudProviderGCP),
		string(CloudProviderAzure),
	}
}

// IsValidCloudProvider reports whether the given string is a known cloud provider.
func IsValidCloudProvider(provider string) bool {
	for _, p := range ValidCloudProviders() {
		if p == provider {
			return true
		}
	}
	return false
}

// IsValidPerformanceTier reports whether the given string is a known performance tier.
func IsValidPerformanceTier(tier string) bool {
	for _, t := range PerformanceTiers {
		if string(t) == tier {
			return true
		}
	}
	return false
}

// FormatBytes returns a human-readable string like "3 GB" or "6.25 GB".
func FormatBytes(bytes uint64) string {
	if bytes >= GB && bytes%GB == 0 {
		return fmt.Sprintf("%d GB", bytes/GB)
	}
	// For fractional GB, show two decimal places and trim trailing zeros.
	gb := float64(bytes) / float64(GB)
	s := fmt.Sprintf("%.2f", gb)
	// Trim trailing zeros after decimal point.
	for s[len(s)-1] == '0' {
		s = s[:len(s)-1]
	}
	if s[len(s)-1] == '.' {
		s = s[:len(s)-1]
	}
	return s + " GB"
}

// FormatMemorySizeList returns a comma-separated, human-readable list
// of memory sizes (e.g. "3 GB, 6.25 GB, 12.5 GB").
func FormatMemorySizeList(sizes []uint64) string {
	if len(sizes) == 0 {
		return "(none)"
	}
	s := ""
	for i, sz := range sizes {
		if i > 0 {
			s += ", "
		}
		s += FormatBytes(sz)
	}
	return s
}
