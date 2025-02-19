package llmserver

import (
    "context"
    "fmt"
    "net/url"
    "strings"

    "github.com/ollama/ollama/api"
)

// Client handles interactions with the Ollama API
type Client struct {
    ollama *api.Client
    cfg    *Config  // Add config field
}

// NewClient creates a new Ollama client
func NewClient(cfg *Config) (*Client, error) {
    baseURL, err := url.Parse(cfg.OllamaURL)
    if err != nil {
        return nil, fmt.Errorf("parse ollama URL: %w", err)
    }
    client := api.NewClient(baseURL, cfg.HTTPClient)
    return &Client{
        ollama: client,
        cfg:    cfg,
    }, nil
}

func (c *Client) AnalyzeNodeData(ctx context.Context, query string, data *NodeData) (*Analysis, error) {
    if data == nil {
        return nil, fmt.Errorf("node data is nil")
    }

    log.Debugf("Starting node analysis with query: %s", query)
    summary := formatNodeSummary(data)
    log.Debugf("Node summary:\n%s", summary)

    req := &api.GenerateRequest{
        Model:  c.cfg.ModelName,
        Prompt: fmt.Sprintf("%s\n\nUser Query: %s", summary, query),
        Stream: ptr(false),
        Options: map[string]interface{}{
            "temperature":     0.7,
            "top_p":          0.9,
            "num_predict":    2048,
            "repeat_penalty": 1.1,
        },
    }

    log.Debugf("Sending request to model %s with prompt:\n%s",
        c.cfg.ModelName, req.Prompt)

    var analysisText strings.Builder
    err := c.ollama.Generate(ctx, req, func(resp api.GenerateResponse) error {
        if resp.Response != "" {
            log.Debugf("Received response chunk: %q", resp.Response)
            analysisText.WriteString(resp.Response)
        }
        return nil
    })

    if err != nil {
        log.Errorf("Analysis generation failed: %v", err)
        return nil, fmt.Errorf("generate analysis: %w", err)
    }

    content := analysisText.String()
    if content == "" {
        log.Errorf("Empty response from model")
        return nil, fmt.Errorf("empty response from model")
    }

    analysis := &Analysis{
        Content: content,
        Data:    extractMetrics(data),
    }

    log.Debugf("Analysis completed successfully (length: %d):\n%s",
        len(analysis.Content), analysis.Content)
    return analysis, nil
}

func formatNodeSummary(data *NodeData) string {
    var summary strings.Builder

    // Add node info section
    if data.NodeInfo != nil {
        summary.WriteString(fmt.Sprintf("Node Analysis for %s\n\n", data.NodeInfo.Alias))
    }

    // Add channel summary with better formatting
    summary.WriteString("Channel Overview:\n")
    for _, ch := range data.Channels {
        alias := ch.PubKeyBytes.String()[:8] // Use first 8 chars of pubkey
        summary.WriteString(fmt.Sprintf("- %d sat channel with %s\n  Local: %d sat | Remote: %d sat | Active: %v\n",
            ch.Capacity, alias, ch.LocalBalance, ch.RemoteBalance, ch.Active))
    }

    // Add enhanced statistics section
    summary.WriteString("\nNode Statistics:\n")
    if data.ChannelStats != nil {
        totalCap := data.ChannelStats.TotalLocalBalance + data.ChannelStats.TotalRemoteBalance
        if totalCap > 0 {
            localRatio := float64(data.ChannelStats.TotalLocalBalance) / float64(totalCap) * 100
            summary.WriteString(fmt.Sprintf("- Balance Distribution: %.1f%% local / %.1f%% remote\n",
                localRatio, 100-localRatio))
        }
        summary.WriteString(fmt.Sprintf("- Active Channels: %d of %d total\n",
            data.ChannelStats.ActiveCount, len(data.Channels)))
    }

    // Add detailed forwarding metrics
    if data.ForwardingStats != nil {
        summary.WriteString("\nForwarding Performance (Last 7 Days):\n")
        summary.WriteString(fmt.Sprintf("- Successful Forwards: %d\n", data.ForwardingStats.SuccessfulCount))
        summary.WriteString(fmt.Sprintf("- Total Volume: %d sat\n", data.ForwardingStats.TotalForwarded/1000))
        summary.WriteString(fmt.Sprintf("- Total Fees Earned: %d sat\n", data.ForwardingStats.TotalFees/1000))

        // Add daily breakdown
        if len(data.ForwardingStats.DailyForwards) > 0 {
            summary.WriteString("\nDaily Forward Counts:\n")
            for day, count := range data.ForwardingStats.DailyForwards {
                summary.WriteString(fmt.Sprintf("- %s: %d forwards\n", day, count))
            }
        }
    }

    // Add analysis guidelines
    summary.WriteString("\nPlease provide a detailed analysis of:\n")
    summary.WriteString("1. Channel balance distribution and rebalancing opportunities\n")
    summary.WriteString("2. Routing performance and fee optimization strategies\n")
    summary.WriteString("3. Peer connectivity and channel health assessment\n")
    summary.WriteString("4. Specific recommendations for improving node performance\n")

    return summary.String()
}

// extractMetrics now includes more detailed statistics
func extractMetrics(data *NodeData) map[string]any {
    metrics := make(map[string]any)

    if data.ChannelStats != nil {
        totalCap := data.ChannelStats.TotalLocalBalance + data.ChannelStats.TotalRemoteBalance
        metrics["total_capacity"] = totalCap
        metrics["active_channels"] = data.ChannelStats.ActiveCount
        metrics["total_channels"] = len(data.Channels)

        if totalCap > 0 {
            metrics["local_balance_ratio"] = float64(data.ChannelStats.TotalLocalBalance) / float64(totalCap)
        }
    }

    if data.ForwardingStats != nil {
        metrics["weekly_forwards"] = data.ForwardingStats.SuccessfulCount
        metrics["total_fees_sat"] = data.ForwardingStats.TotalFees / 1000 // Convert to sats
        metrics["total_volume_sat"] = data.ForwardingStats.TotalForwarded / 1000
        metrics["daily_forwards"] = data.ForwardingStats.DailyForwards
    }

    if data.PeerStats != nil {
        metrics["connected_peers"] = data.PeerStats.ConnectedCount
        metrics["total_peers"] = data.PeerStats.ConnectedCount + data.PeerStats.DisconnectedCount
        metrics["inbound_peers"] = data.PeerStats.InboundCount
        metrics["outbound_peers"] = data.PeerStats.OutboundCount
    }

    return metrics
}

// Helper functions
func ptr(b bool) *bool { return &b }
