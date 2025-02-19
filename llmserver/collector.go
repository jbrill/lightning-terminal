package llmserver

import (
    "context"
    "fmt"
    "time"

    "github.com/lightninglabs/lndclient"
)

// NodeCollector handles collecting node data from LND
type NodeCollector struct {
    lndClient lndclient.LightningClient
}

func NewNodeCollector(lndClient lndclient.LightningClient) *NodeCollector {
    return &NodeCollector{lndClient: lndClient}
}

func (c *NodeCollector) CollectNodeData(ctx context.Context) (*NodeData, error) {
    data := &NodeData{}

    info, err := c.lndClient.GetInfo(ctx)
    if err != nil {
        log.Errorf("Failed to get node info: %v", err)
        return nil, fmt.Errorf("error getting node info: %w", err)
    }
    data.NodeInfo = info

    channels, err := c.lndClient.ListChannels(ctx, true, true)
    if err != nil {
        log.Errorf("Failed to get channels: %v", err)
        return nil, fmt.Errorf("error getting channels: %w", err)
    }
    data.Channels = channels

    // Convert amounts using Int64()
    var totalLocal, totalRemote int64
    activeCount := 0
    for _, channel := range channels {
        totalLocal += int64(channel.LocalBalance)
        totalRemote += int64(channel.RemoteBalance)
        if channel.Active {
            activeCount++
        }
    }

    data.ChannelStats = &ChannelStats{
        TotalLocalBalance:  totalLocal,
        TotalRemoteBalance: totalRemote,
        ActiveCount:        activeCount,
    }

    now := time.Now()
    data.ForwardingStats, err = c.collectForwardingStats(ctx, now)
    if err != nil {
        log.Errorf("Failed to get forwarding stats: %v", err)
        return nil, fmt.Errorf("error getting forwarding stats: %w", err)
    }

    data.PeerStats, err = c.collectPeerStats(ctx)
    if err != nil {
        log.Errorf("Failed to collect peer stats: %v", err)
        return nil, fmt.Errorf("error getting peer stats: %w", err)
    }

    return data, nil
}

func (c *NodeCollector) collectForwardingStats(ctx context.Context, now time.Time) (*ForwardingStats, error) {
    stats := &ForwardingStats{
        WeeklyForwards:  make(map[string]uint64),
        DailyForwards:   make(map[string]uint64),
        TotalFees:       0,
        TotalForwarded:  0,
        SuccessfulCount: 0,
    }

    weekAgo := now.Add(-7 * 24 * time.Hour)
    forwards, err := c.lndClient.ForwardingHistory(ctx, lndclient.ForwardingHistoryRequest{
        StartTime: weekAgo,
        EndTime:   now,
        MaxEvents: 1000,
    })
    if err != nil {
        return nil, err
    }

    for _, event := range forwards.Events {
        day := event.Timestamp.Format("2006-01-02")
        stats.DailyForwards[day]++
        // Convert to int64 since we're dealing with msat
        stats.TotalFees += int64(event.FeeMsat)
        stats.TotalForwarded += int64(0) // TODO: get actual forwarded amount
        stats.SuccessfulCount++
    }

    return stats, nil
}

func (c *NodeCollector) collectPeerStats(ctx context.Context) (*PeerStats, error) {
    stats := &PeerStats{
        ConnectedCount:    0,
        DisconnectedCount: 0,
        TotalBytes:        0,
    }

    peers, err := c.lndClient.ListPeers(ctx)
    if err != nil {
        return nil, fmt.Errorf("list peers: %w", err)
    }

    for _, peer := range peers {
        // Check online status via active channels
        hasActiveChannels := false
        channels, err := c.lndClient.ListChannels(ctx, false, false)
        if err == nil {
            for _, ch := range channels {
                if ch.PubKeyBytes == peer.Pubkey && ch.Active {
                    hasActiveChannels = true
                    break
                }
            }
        }

        if hasActiveChannels {
            stats.ConnectedCount++
        } else {
            stats.DisconnectedCount++
        }

        if peer.Inbound {
            stats.InboundCount++
        } else {
            stats.OutboundCount++
        }

        // Since we can't get direct bytes, estimate based on channels
        stats.TotalBytes += uint64(len(channels)) * 1000 // rough estimate
    }

    return stats, nil
}
