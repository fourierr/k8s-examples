package main

import (
	flag "github.com/spf13/pflag"
	"k8s.io/apimachinery/pkg/util/runtime"
	"k8s.io/apiserver/pkg/util/feature"
	"k8s.io/component-base/featuregate"
	"time"
)

const (
	// DeprecatedPolicySpec enable the use of deprecated policy spec
	DeprecatedPolicySpec featuregate.Feature = "DeprecatedPolicySpec"
	// LegacyObjectTypeIdentifier enable the use of legacy object type identifier for selecting ref-object
	LegacyObjectTypeIdentifier featuregate.Feature = "LegacyObjectTypeIdentifier"
	// DeprecatedObjectLabelSelector enable the use of deprecated object label selector for selecting ref-object
	DeprecatedObjectLabelSelector featuregate.Feature = "DeprecatedObjectLabelSelector"
	// GzipResourceTracker enables the gzip compression for ResourceTracker. It can be useful if you have large
	// application that needs to dispatch lots of resources or large resources (like CRD or huge ConfigMap),
	// which at the cost of slower processing speed due to the extra overhead for compression and decompression.
	GzipResourceTracker featuregate.Feature = "GzipResourceTracker"
	// ZstdResourceTracker enables the zstd compression for ResourceTracker.
	// Refer to GzipResourceTracker for its use-cases. It is much faster and more
	// efficient than gzip, about 2x faster and compresses to smaller size.
	// If you are dealing with very large ResourceTrackers (1MB or so), it should
	// have almost NO performance penalties compared to no compression at all.
	// If dealing with smaller ResourceTrackers (10KB - 1MB), the performance
	// penalties are minimal.
	ZstdResourceTracker featuregate.Feature = "ZstdResourceTracker"
	// GzipApplicationRevision enables the gzip compression for ApplicationRevision. It can be useful if you have large
	// application that needs to dispatch lots of resources or large resources (like CRD or huge ConfigMap),
	// which at the cost of slower processing speed due to the extra overhead for compression and decompression.
	GzipApplicationRevision featuregate.Feature = "GzipApplicationRevision"
	// ZstdApplicationRevision enables the zstd compression for ApplicationRevision.
	// Refer to GzipApplicationRevision for its use-cases. It is much faster and more
	// efficient than gzip, about 2x faster and compresses to smaller size.
	// If you are dealing with very large ResourceTrackers (1MB or so), it should
	// have almost NO performance penalties compared to no compression at all.
	// If dealing with smaller ResourceTrackers (10KB - 1MB), the performance
	// penalties are minimal.
	ZstdApplicationRevision featuregate.Feature = "ZstdApplicationRevision"
)

var defaultFeatureGates = map[featuregate.Feature]featuregate.FeatureSpec{
	DeprecatedPolicySpec:          {Default: false, PreRelease: featuregate.Alpha},
	LegacyObjectTypeIdentifier:    {Default: false, PreRelease: featuregate.Alpha},
	DeprecatedObjectLabelSelector: {Default: false, PreRelease: featuregate.Alpha},
	GzipResourceTracker:           {Default: false, PreRelease: featuregate.Alpha},
	ZstdResourceTracker:           {Default: true, PreRelease: featuregate.Alpha},
	GzipApplicationRevision:       {Default: false, PreRelease: featuregate.Alpha},
	ZstdApplicationRevision:       {Default: true, PreRelease: featuregate.Alpha},
}

func init() {
	runtime.Must(feature.DefaultMutableFeatureGate.Add(defaultFeatureGates))
}

func main() {
	var leaseDuration time.Duration
	var renewDeadline time.Duration
	var retryPeriod time.Duration
	flag.DurationVar(&leaseDuration, "leader-election-lease-duration", 15*time.Second,
		"The duration that non-leader candidates will wait to force acquire leadership")
	flag.DurationVar(&renewDeadline, "leader-election-renew-deadline", 10*time.Second,
		"The duration that the acting controlplane will retry refreshing leadership before giving up")
	flag.DurationVar(&retryPeriod, "leader-election-retry-period", 2*time.Second,
		"The duration the LeaderElector clients should wait between tries of actions")
	feature.DefaultMutableFeatureGate.AddFlag(flag.CommandLine)
	flag.Parse()
	time.Sleep(5 * time.Minute)
}
