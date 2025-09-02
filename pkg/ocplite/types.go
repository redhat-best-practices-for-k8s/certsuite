package ocplite

// ClusterOperator minimal representation needed by certsuite
type ClusterOperator struct {
	Name   string                `json:"name"`
	Status ClusterOperatorStatus `json:"status"`
}

type ClusterOperatorStatus struct {
	Conditions []ClusterOperatorStatusCondition `json:"conditions"`
	Versions   []OperandVersion                 `json:"versions"`
}

type ClusterOperatorStatusCondition struct {
	Type   string `json:"type"`
	Status string `json:"status"`
}

type OperandVersion struct {
	Name    string `json:"name"`
	Version string `json:"version"`
}

const (
	OperatorAvailable   = "Available"
	OperatorProgressing = "Progressing"
	OperatorDegraded    = "Degraded"
)

// APIRequestCount minimal representation used in observability tests
type APIRequestCount struct {
	Name   string                `json:"name"`
	Status APIRequestCountStatus `json:"status"`
}

type APIRequestCountStatus struct {
	RemovedInRelease string    `json:"removedInRelease"`
	Last24h          []Last24h `json:"last24h"`
}

type Last24h struct {
	ByNode []ByNode `json:"byNode"`
}

type ByNode struct {
	ByUser []PerUserAPIRequestCount `json:"byUser"`
}

type PerUserAPIRequestCount struct {
	UserName string `json:"userName"`
}
