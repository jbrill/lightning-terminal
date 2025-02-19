package llmserver

import (
    "github.com/lightninglabs/lndclient"
)

// Analysis represents the analysis response
type Analysis struct {
    Content string         `json:"content"`  // Free-form analysis text
    Data    map[string]any `json:"data"`    // Optional structured data
    Done    bool          `json:"done"`     // Indicates if this is the final chunk
}

// AnalysisCallback is a function that receives streaming analysis updates
type AnalysisCallback func(*Analysis) error

// Add new types for enhanced data collection
type ChannelStats struct {
    TotalLocalBalance  int64
    TotalRemoteBalance int64
    ActiveCount        int
}

type ForwardingStats struct {
    WeeklyForwards   map[string]uint64
    DailyForwards    map[string]uint64
    TotalFees        int64
    TotalForwarded   int64
    SuccessfulCount  uint64
}

type PeerStats struct {
    ConnectedCount    int
    DisconnectedCount int
    InboundCount     int
    OutboundCount    int
    TotalBytes       uint64
}

// NodeData includes all node statistics
type NodeData struct {
    NodeInfo        *lndclient.Info
    Channels        []lndclient.ChannelInfo
    Forwards        []lndclient.ForwardingEvent
    ChannelStats    *ChannelStats
    ForwardingStats *ForwardingStats
    PeerStats       *PeerStats
}
