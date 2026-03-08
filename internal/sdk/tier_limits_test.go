package sdk

import (
	"testing"
)

func TestIsSupportedShardSize(t *testing.T) {
	tests := []struct {
		name      string
		tier      PerformanceTier
		sizeBytes uint64
		provider  CloudProvider
		want      bool
	}{
		// Dev tier
		{name: "dev 3GB AWS", tier: PerformanceTierDev, sizeBytes: 3 * GB, provider: CloudProviderAWS, want: true},
		{name: "dev 3GB GCP", tier: PerformanceTierDev, sizeBytes: 3 * GB, provider: CloudProviderGCP, want: true},
		{name: "dev 3GB Azure", tier: PerformanceTierDev, sizeBytes: 3 * GB, provider: CloudProviderAzure, want: true},
		{name: "dev 6GB AWS rejected", tier: PerformanceTierDev, sizeBytes: 6 * GB, provider: CloudProviderAWS, want: false},
		{name: "dev 30GB AWS rejected", tier: PerformanceTierDev, sizeBytes: 30 * GB, provider: CloudProviderAWS, want: false},
		{name: "dev 12.5GB rejected", tier: PerformanceTierDev, sizeBytes: 12*GB + 500*MB, provider: CloudProviderAWS, want: false},

		// Standard tier
		{name: "standard 12.5GB AWS", tier: PerformanceTierStandard, sizeBytes: 12*GB + 500*MB, provider: CloudProviderAWS, want: true},
		{name: "standard 25GB GCP", tier: PerformanceTierStandard, sizeBytes: 25 * GB, provider: CloudProviderGCP, want: true},
		{name: "standard 300GB AWS rejected", tier: PerformanceTierStandard, sizeBytes: 300 * GB, provider: CloudProviderAWS, want: false},
		{name: "standard 300GB GCP", tier: PerformanceTierStandard, sizeBytes: 300 * GB, provider: CloudProviderGCP, want: true},
		{name: "standard 500GB Azure rejected", tier: PerformanceTierStandard, sizeBytes: 500 * GB, provider: CloudProviderAzure, want: false},
		{name: "standard 500GB AWS", tier: PerformanceTierStandard, sizeBytes: 500 * GB, provider: CloudProviderAWS, want: true},
		{name: "standard 3GB rejected", tier: PerformanceTierStandard, sizeBytes: 3 * GB, provider: CloudProviderAWS, want: false},
		{name: "standard 30GB rejected", tier: PerformanceTierStandard, sizeBytes: 30 * GB, provider: CloudProviderAWS, want: false},

		// Enhanced tier
		{name: "enhanced 6.25GB AWS", tier: PerformanceTierEnhanced, sizeBytes: 6*GB + 250*MB, provider: CloudProviderAWS, want: true},
		{name: "enhanced 6.25GB Azure rejected", tier: PerformanceTierEnhanced, sizeBytes: 6*GB + 250*MB, provider: CloudProviderAzure, want: false},
		{name: "enhanced 6.5GB Azure", tier: PerformanceTierEnhanced, sizeBytes: 6*GB + 500*MB, provider: CloudProviderAzure, want: true},
		{name: "enhanced 6.5GB AWS rejected", tier: PerformanceTierEnhanced, sizeBytes: 6*GB + 500*MB, provider: CloudProviderAWS, want: false},
		{name: "enhanced 250GB GCP", tier: PerformanceTierEnhanced, sizeBytes: 250 * GB, provider: CloudProviderGCP, want: true},
		{name: "enhanced 250GB AWS rejected", tier: PerformanceTierEnhanced, sizeBytes: 250 * GB, provider: CloudProviderAWS, want: false},
		{name: "enhanced 30GB rejected", tier: PerformanceTierEnhanced, sizeBytes: 30 * GB, provider: CloudProviderGCP, want: false},

		// Extreme tier
		{name: "extreme 6.25GB AWS", tier: PerformanceTierExtreme, sizeBytes: 6*GB + 250*MB, provider: CloudProviderAWS, want: true},
		{name: "extreme 200GB Azure", tier: PerformanceTierExtreme, sizeBytes: 200 * GB, provider: CloudProviderAzure, want: true},
		{name: "extreme 150GB GCP", tier: PerformanceTierExtreme, sizeBytes: 150 * GB, provider: CloudProviderGCP, want: true},
		{name: "extreme 150GB AWS rejected", tier: PerformanceTierExtreme, sizeBytes: 150 * GB, provider: CloudProviderAWS, want: false},
		{name: "extreme 300GB rejected", tier: PerformanceTierExtreme, sizeBytes: 300 * GB, provider: CloudProviderAWS, want: false},

		// BYOC tier
		{name: "byoc 6.25GB AWS", tier: PerformanceTierBYOC, sizeBytes: 6*GB + 250*MB, provider: CloudProviderAWS, want: true},
		{name: "byoc 6.25GB Azure rejected", tier: PerformanceTierBYOC, sizeBytes: 6*GB + 250*MB, provider: CloudProviderAzure, want: false},
		{name: "byoc 300GB GCP", tier: PerformanceTierBYOC, sizeBytes: 300 * GB, provider: CloudProviderGCP, want: true},
		{name: "byoc 300GB AWS rejected", tier: PerformanceTierBYOC, sizeBytes: 300 * GB, provider: CloudProviderAWS, want: false},

		// Unknown tier
		{name: "unknown tier rejected", tier: "nonexistent", sizeBytes: 3 * GB, provider: CloudProviderAWS, want: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := IsSupportedShardSize(tt.tier, tt.sizeBytes, tt.provider)
			if got != tt.want {
				t.Errorf("IsSupportedShardSize(%q, %d, %q) = %v, want %v",
					tt.tier, tt.sizeBytes, tt.provider, got, tt.want)
			}
		})
	}
}

func TestSupportedSizesForTier(t *testing.T) {
	tests := []struct {
		name     string
		tier     PerformanceTier
		provider CloudProvider
		wantLen  int
	}{
		{name: "dev AWS has 1 size", tier: PerformanceTierDev, provider: CloudProviderAWS, wantLen: 1},
		{name: "dev GCP has 1 size", tier: PerformanceTierDev, provider: CloudProviderGCP, wantLen: 1},
		{name: "standard AWS", tier: PerformanceTierStandard, provider: CloudProviderAWS, wantLen: 8},     // 12.5, 25, 50, 100, 200, 400, 500, 600
		{name: "standard Azure", tier: PerformanceTierStandard, provider: CloudProviderAzure, wantLen: 7}, // 12.5, 25, 50, 100, 200, 300, 400
		{name: "unknown tier", tier: "nonexistent", provider: CloudProviderAWS, wantLen: 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			sizes := SupportedSizesForTier(tt.tier, tt.provider)
			if len(sizes) != tt.wantLen {
				t.Errorf("SupportedSizesForTier(%q, %q) returned %d sizes, want %d; sizes=%v",
					tt.tier, tt.provider, len(sizes), tt.wantLen, sizes)
			}
		})
	}
}

func TestFormatBytes(t *testing.T) {
	tests := []struct {
		bytes uint64
		want  string
	}{
		{bytes: 3 * GB, want: "3 GB"},
		{bytes: 6*GB + 250*MB, want: "6.25 GB"},
		{bytes: 6*GB + 500*MB, want: "6.5 GB"},
		{bytes: 12*GB + 500*MB, want: "12.5 GB"},
		{bytes: 25 * GB, want: "25 GB"},
		{bytes: 100 * GB, want: "100 GB"},
		{bytes: 200 * GB, want: "200 GB"},
		{bytes: 30 * GB, want: "30 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.want, func(t *testing.T) {
			got := FormatBytes(tt.bytes)
			if got != tt.want {
				t.Errorf("FormatBytes(%d) = %q, want %q", tt.bytes, got, tt.want)
			}
		})
	}
}

func TestFormatMemorySizeList(t *testing.T) {
	tests := []struct {
		name  string
		sizes []uint64
		want  string
	}{
		{name: "empty", sizes: nil, want: "(none)"},
		{name: "single", sizes: []uint64{3 * GB}, want: "3 GB"},
		{name: "multiple", sizes: []uint64{3 * GB, 6*GB + 250*MB, 12*GB + 500*MB}, want: "3 GB, 6.25 GB, 12.5 GB"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := FormatMemorySizeList(tt.sizes)
			if got != tt.want {
				t.Errorf("FormatMemorySizeList(%v) = %q, want %q", tt.sizes, got, tt.want)
			}
		})
	}
}

func TestIsValidCloudProvider(t *testing.T) {
	if !IsValidCloudProvider("aws") {
		t.Error("expected aws to be valid")
	}
	if !IsValidCloudProvider("gcp") {
		t.Error("expected gcp to be valid")
	}
	if !IsValidCloudProvider("azure") {
		t.Error("expected azure to be valid")
	}
}

func TestIsValidPerformanceTier(t *testing.T) {
	for _, tier := range PerformanceTiers {
		if !IsValidPerformanceTier(string(tier)) {
			t.Errorf("expected %q to be valid", tier)
		}
	}
	if IsValidPerformanceTier("mega") {
		t.Error("expected mega to be invalid")
	}
}
