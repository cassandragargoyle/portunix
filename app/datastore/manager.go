package datastore

import (
	"context"
	"fmt"
	"sync"
	"time"

	"portunix.cz/app/plugins"
	"portunix.cz/app/plugins/manager"
	"portunix.cz/app/plugins/proto"
)

// Manager manages multiple datastore instances and routes operations
type Manager struct {
	config        *Config
	datastores    map[string]DatastoreInterface
	pluginManager *manager.Manager
	mu            sync.RWMutex
	initialized   bool
}

// NewManager creates a new datastore manager
func NewManager(config *Config, pluginManager *manager.Manager) (*Manager, error) {
	if err := ValidateConfig(config); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	dsManager := &Manager{
		config:        config,
		datastores:    make(map[string]DatastoreInterface),
		pluginManager: pluginManager,
	}

	return dsManager, nil
}

// Initialize initializes all configured datastores
func (m *Manager) Initialize(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if m.initialized {
		return nil
	}

	// Initialize all plugins mentioned in routes
	pluginsToInit := make(map[string]bool)
	for _, route := range m.config.Routes {
		pluginsToInit[route.Plugin] = true
	}

	for pluginName := range pluginsToInit {
		if err := m.initializePlugin(ctx, pluginName); err != nil {
			return fmt.Errorf("failed to initialize plugin '%s': %w", pluginName, err)
		}
	}

	m.initialized = true
	return nil
}

// initializePlugin initializes a specific datastore plugin
func (m *Manager) initializePlugin(ctx context.Context, pluginName string) error {
	pluginConfig, exists := m.config.Plugins[pluginName]
	if !exists {
		return fmt.Errorf("configuration for plugin '%s' not found", pluginName)
	}

	// Handle built-in file plugin
	if pluginName == "file-plugin" {
		fileStore := NewFileDatastore()
		if err := fileStore.Initialize(ctx, pluginConfig.Settings); err != nil {
			return fmt.Errorf("failed to initialize file datastore: %w", err)
		}

		m.datastores[pluginName] = fileStore
		return nil
	}

	// Handle external datastore plugins via gRPC
	if m.pluginManager != nil {
		// Check if plugin is installed and running
		plugin, err := m.pluginManager.GetPlugin(pluginName)
		if err != nil {
			return fmt.Errorf("datastore plugin '%s' not found: %w", pluginName, err)
		}

		// Check if plugin is a datastore plugin
		if !m.isDatastorePlugin(&plugin) {
			return fmt.Errorf("plugin '%s' is not a datastore plugin", pluginName)
		}

		// Start plugin if not running
		if plugin.Status != plugins.PluginStatusRunning {
			if err := m.pluginManager.StartPlugin(pluginName); err != nil {
				return fmt.Errorf("failed to start datastore plugin '%s': %w", pluginName, err)
			}
		}

		// Create gRPC client for the plugin
		grpcClient, err := m.createDatastorePluginClient(ctx, &plugin)
		if err != nil {
			return fmt.Errorf("failed to create gRPC client for plugin '%s': %w", pluginName, err)
		}

		// Create plugin datastore wrapper
		pluginStore := NewPluginDatastore(grpcClient, pluginName)
		if err := pluginStore.Initialize(ctx, pluginConfig.Settings); err != nil {
			return fmt.Errorf("failed to initialize plugin datastore '%s': %w", pluginName, err)
		}

		m.datastores[pluginName] = pluginStore
		return nil
	}

	return fmt.Errorf("plugin '%s' not supported - plugin manager not available", pluginName)
}

// isDatastorePlugin checks if a plugin is a datastore plugin
func (m *Manager) isDatastorePlugin(plugin *plugins.PluginInfo) bool {
	// Check if plugin has database access capability
	return plugin.Capabilities.DatabaseAccess
}

// createDatastorePluginClient creates a gRPC client for a datastore plugin
func (m *Manager) createDatastorePluginClient(ctx context.Context, plugin *plugins.PluginInfo) (proto.DatastorePluginServiceClient, error) {
	// TODO: Implement actual gRPC client creation
	// This would connect to the plugin's gRPC server using the plugin's port
	// For now, return nil as placeholder
	return nil, fmt.Errorf("gRPC client creation not yet implemented")
}

// Store stores data using the appropriate datastore based on routing rules
func (m *Manager) Store(ctx context.Context, key string, value interface{}, metadata map[string]interface{}) error {
	route := m.config.FindRoute(key)
	if route == nil {
		return fmt.Errorf("no route found for key: %s", key)
	}

	datastore, err := m.getDatastore(route.Plugin)
	if err != nil {
		return fmt.Errorf("failed to get datastore for plugin '%s': %w", route.Plugin, err)
	}

	// Merge route config with metadata
	if metadata == nil {
		metadata = make(map[string]interface{})
	}
	metadata["route_name"] = route.Name
	metadata["plugin"] = route.Plugin
	
	// Add route-specific config to metadata
	for k, v := range route.Config {
		metadata["route_"+k] = v
	}

	return datastore.Store(ctx, key, value, metadata)
}

// Retrieve retrieves data using the appropriate datastore
func (m *Manager) Retrieve(ctx context.Context, key string, filter map[string]interface{}) (interface{}, error) {
	route := m.config.FindRoute(key)
	if route == nil {
		return nil, fmt.Errorf("no route found for key: %s", key)
	}

	datastore, err := m.getDatastore(route.Plugin)
	if err != nil {
		return nil, fmt.Errorf("failed to get datastore for plugin '%s': %w", route.Plugin, err)
	}

	return datastore.Retrieve(ctx, key, filter)
}

// Query queries data across appropriate datastores
func (m *Manager) Query(ctx context.Context, criteria QueryCriteria) ([]QueryResult, error) {
	// Determine which datastores to query based on criteria
	var results []QueryResult
	
	// For now, query all datastores that match the collection pattern
	// TODO: Optimize to only query relevant datastores
	
	pluginsToQuery := make(map[string]bool)
	for _, route := range m.config.Routes {
		if criteria.Collection == "" || matchPattern(route.Pattern, criteria.Collection) {
			pluginsToQuery[route.Plugin] = true
		}
	}

	for pluginName := range pluginsToQuery {
		datastore, err := m.getDatastore(pluginName)
		if err != nil {
			continue // Skip unavailable datastores
		}

		pluginResults, err := datastore.Query(ctx, criteria)
		if err != nil {
			// Log error but continue with other datastores
			continue
		}

		results = append(results, pluginResults...)
	}

	return results, nil
}

// Delete deletes data using the appropriate datastore
func (m *Manager) Delete(ctx context.Context, key string) error {
	route := m.config.FindRoute(key)
	if route == nil {
		return fmt.Errorf("no route found for key: %s", key)
	}

	datastore, err := m.getDatastore(route.Plugin)
	if err != nil {
		return fmt.Errorf("failed to get datastore for plugin '%s': %w", route.Plugin, err)
	}

	return datastore.Delete(ctx, key)
}

// List lists keys matching a pattern across appropriate datastores
func (m *Manager) List(ctx context.Context, pattern string) ([]string, error) {
	var allKeys []string
	
	// Find all plugins that could contain keys matching the pattern
	pluginsToSearch := make(map[string]bool)
	for _, route := range m.config.Routes {
		if matchPattern(route.Pattern, pattern) || matchPattern(pattern, route.Pattern) {
			pluginsToSearch[route.Plugin] = true
		}
	}

	for pluginName := range pluginsToSearch {
		datastore, err := m.getDatastore(pluginName)
		if err != nil {
			continue // Skip unavailable datastores
		}

		keys, err := datastore.List(ctx, pattern)
		if err != nil {
			continue // Skip errors, continue with other datastores
		}

		allKeys = append(allKeys, keys...)
	}

	return allKeys, nil
}

// Health returns health status of all datastores
func (m *Manager) Health(ctx context.Context) (map[string]*HealthStatus, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	healthMap := make(map[string]*HealthStatus)
	
	for pluginName, datastore := range m.datastores {
		health, err := datastore.Health(ctx)
		if err != nil {
			healthMap[pluginName] = &HealthStatus{
				Healthy:   false,
				Status:    "error",
				Message:   err.Error(),
				LastCheck: time.Now(),
			}
		} else {
			healthMap[pluginName] = health
		}
	}

	return healthMap, nil
}

// Stats returns statistics from all datastores
func (m *Manager) Stats(ctx context.Context) (map[string]*Stats, error) {
	m.mu.RLock()
	defer m.mu.RUnlock()

	statsMap := make(map[string]*Stats)
	
	for pluginName, datastore := range m.datastores {
		stats, err := datastore.Stats(ctx)
		if err != nil {
			continue // Skip errors
		}
		statsMap[pluginName] = stats
	}

	return statsMap, nil
}

// GetDatastoreInfo returns information about available datastores
func (m *Manager) GetDatastoreInfo(ctx context.Context) (map[string]*DatastoreInfo, error) {
	info := make(map[string]*DatastoreInfo)
	
	// For now, only file datastore is supported
	info["file-plugin"] = &DatastoreInfo{
		Name:        "File Datastore",
		Type:        DatastoreTypeFile,
		Version:     "1.0.0",
		Description: "Local file-based datastore with YAML/JSON support",
		Capabilities: DatastoreCapabilities{
			SupportsTransactions: false,
			SupportsIndexing:     false,
			SupportsAggregation:  false,
			SupportsFullText:     false,
			SupportedDataTypes:   []string{"json", "yaml", "text", "binary"},
			MaxKeySize:           1024,
			MaxValueSize:         100 * 1024 * 1024, // 100MB
		},
	}

	return info, nil
}

// Close closes all datastore connections
func (m *Manager) Close(ctx context.Context) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	var errors []error
	for pluginName, datastore := range m.datastores {
		if err := datastore.Close(ctx); err != nil {
			errors = append(errors, fmt.Errorf("failed to close %s: %w", pluginName, err))
		}
	}

	if len(errors) > 0 {
		return fmt.Errorf("errors while closing datastores: %v", errors)
	}

	m.initialized = false
	return nil
}

// getDatastore gets a datastore instance, initializing if necessary
func (m *Manager) getDatastore(pluginName string) (DatastoreInterface, error) {
	m.mu.RLock()
	datastore, exists := m.datastores[pluginName]
	m.mu.RUnlock()

	if exists {
		return datastore, nil
	}

	// Try to initialize the plugin if not found
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m.mu.Lock()
	defer m.mu.Unlock()

	// Double-check after acquiring write lock
	if datastore, exists := m.datastores[pluginName]; exists {
		return datastore, nil
	}

	if err := m.initializePlugin(ctx, pluginName); err != nil {
		return nil, err
	}

	return m.datastores[pluginName], nil
}

// ReloadConfig reloads the configuration and reinitializes datastores
func (m *Manager) ReloadConfig(ctx context.Context, newConfig *Config) error {
	if err := ValidateConfig(newConfig); err != nil {
		return fmt.Errorf("invalid configuration: %w", err)
	}

	m.mu.Lock()
	defer m.mu.Unlock()

	// Close existing datastores
	for _, datastore := range m.datastores {
		datastore.Close(ctx)
	}

	// Update config and reinitialize
	m.config = newConfig
	m.datastores = make(map[string]DatastoreInterface)
	m.initialized = false

	return m.Initialize(ctx)
}