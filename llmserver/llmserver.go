package llmserver

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/lightninglabs/lightning-terminal/litrpc"
	"github.com/lightninglabs/lndclient"
	"github.com/ollama/ollama/api"
)

const (
	// DefaultOllamaURL is the default URL for the Ollama server.
	DefaultOllamaURL = "http://host.docker.internal:11434"

	// DefaultModelName is the default model name to use.
	DefaultModelName = "lit-analysis"

	// DefaultTimeout is the default timeout for Ollama API requests
	DefaultTimeout = 300 * time.Second
)

// Config holds the configuration for the LLM server
type Config struct {
	// OllamaURL is the URL of the Ollama server
	OllamaURL string

	// ModelName is the name of the model to use
	ModelName string

	// HTTPClient is the HTTP client to use
	HTTPClient *http.Client

	// Timeout is the timeout for Ollama API requests
	Timeout time.Duration

	// Disable is a flag to disable the LLM server
	Disable bool
}

// Server implements the LLM gRPC service.
type Server struct {
	litrpc.UnimplementedLLMServer
	cfg       *Config
	client    *Client
	collector *NodeCollector
	manager   *Manager
	ready     atomic.Bool
}

// NewServer creates a new LLM server instance.
func NewServer(cfg *Config, lndClient lndclient.LightningClient) (*Server, error) {
	// Set default timeout if not specified
	if cfg.Timeout == 0 {
		cfg.Timeout = DefaultTimeout
	}

	// Create HTTP client with timeout if not provided
	if cfg.HTTPClient == nil {
		cfg.HTTPClient = &http.Client{
			Timeout: cfg.Timeout,
		}
	}

	// Use the server's config for the client instead of creating a new one
	client, err := NewClient(cfg)
	if err != nil {
		return nil, fmt.Errorf("create client: %w", err)
	}

	return &Server{
		cfg:       cfg,
		client:    client,
		manager:   NewManager(),
		collector: NewNodeCollector(lndClient),
	}, nil
}

// Start starts the LLM server.
func (s *Server) Start(ctx context.Context) error {
	log.Infof("Starting LLM server...")

	// If LLM is disabled, return early without error
	if s.cfg.Disable {
		log.Infof("LLM server disabled, skipping start")
		return nil
	}

	if err := s.manager.Start(ctx, s.cfg); err != nil {
		// If model not found, disable the server but don't return error
		if strings.Contains(err.Error(), "model") && strings.Contains(err.Error(), "not found") {
			log.Warnf("Model not found, disabling LLM server: %v", err)
			s.cfg.Disable = true
			return nil
		}

		log.Errorf("Failed to start LLM server: %v", err)
		return err
	}

	s.ready.Store(true)
	log.Infof("LLM server started successfully")
	return nil
}

// Stop stops the LLM server.
func (s *Server) Stop() error {
	s.manager.stopped.Do(func() {
		log.Infof("Stopping LLM server...")
		s.manager.ready.Store(false)
		log.Infof("LLM server stopped")
	})
	return nil
}

// Manager handles the server lifecycle.
type Manager struct {
	ready   atomic.Bool
	started sync.Once
	stopped sync.Once
}

func NewManager() *Manager {
	return &Manager{}
}

func (m *Manager) Start(ctx context.Context, cfg *Config) error {
	var startErr error
	m.started.Do(func() {
		if err := m.checkDependencies(ctx, cfg); err != nil {
			startErr = err
			return
		}
		m.ready.Store(true)
	})
	return startErr
}

func (m *Manager) checkDependencies(ctx context.Context, cfg *Config) error {
	baseURL, err := url.Parse(cfg.OllamaURL)
	if err != nil {
		return fmt.Errorf("parse ollama URL: %w", err)
	}

	ollamaClient := api.NewClient(baseURL, cfg.HTTPClient)

	// Retry health check with backoff
	maxRetries := 5
	for i := 0; i < maxRetries; i++ {
		// Check server health
		if err := ollamaClient.Heartbeat(ctx); err != nil {
			log.Warnf("Ollama server not available (attempt %d/%d): %v",
				i+1, maxRetries, err)
			if i == maxRetries-1 {
				return fmt.Errorf("ollama server not available after %d attempts: %w",
					maxRetries, err)
			}
			// Exponential backoff
			time.Sleep(time.Second * time.Duration(1<<uint(i)))
			continue
		}
		break
	}

	// Check if model exists
	list, err := ollamaClient.List(ctx)
	if err != nil {
		return fmt.Errorf("list models: %w", err)
	}

	log.Debugf("Available models: %v", list.Models)

	modelExists := false
	searchName := strings.ToLower(cfg.ModelName)
	for _, m := range list.Models {
		// Compare case-insensitive and with/without :latest
		modelName := strings.ToLower(strings.TrimSuffix(m.Name, ":latest"))
		if modelName == searchName {
			modelExists = true
			break
		}
	}

	if !modelExists {
		return fmt.Errorf("model %q not found, please pull or create it first",
			cfg.ModelName)
	}

	return nil
}

// AnalyzeNode implements the LLM.AnalyzeNode RPC.
func (s *Server) AnalyzeNode(req *litrpc.AnalyzeNodeRequest,
	stream litrpc.LLM_AnalyzeNodeServer) error {

	// Check if server is disabled
	if s.cfg.Disable {
		return fmt.Errorf("LLM server is disabled, model not available")
	}

	if !s.ready.Load() {
		log.Errorf("LLM server not ready")
		return fmt.Errorf("LLM server not ready")
	}

	// Create a new context with extended timeout for streaming
	ctx, cancel := context.WithTimeout(stream.Context(), s.cfg.Timeout)
	defer cancel()

	log.Infof("Analyzing node with query: %s", req.Query)

	// Collect node data from LND.
	log.Debugf("Collecting node data...")
	nodeData, err := s.collector.CollectNodeData(ctx)
	if err != nil {
		log.Errorf("Error collecting node data: %v", err)
		return fmt.Errorf("error collecting node data: %w", err)
	}
	log.Debugf("Node data collected successfully")

	// Create streaming callback with heartbeat
	callback := func(analysis *Analysis) error {
		// Send periodic heartbeat to keep connection alive
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			resp := &litrpc.AnalyzeNodeResponse{
				Analysis: analysis.Content,
				Done:     analysis.Done,
			}
			return stream.Send(resp)
		}
	}

	// Run streaming analysis
	log.Debugf("Running LLM analysis...")
	if err := s.client.AnalyzeNodeData(ctx, req.Query, nodeData, callback); err != nil {
		if err == context.DeadlineExceeded {
			log.Errorf("Analysis timed out after %v", s.cfg.Timeout)
			return fmt.Errorf("analysis timed out after %v", s.cfg.Timeout)
		}
		log.Errorf("Error running analysis: %v", err)
		return err
	}
	log.Debugf("Analysis completed successfully")

	return nil
}
