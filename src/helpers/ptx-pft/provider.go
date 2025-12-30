package main

// FeedbackItem represents a single feedback entry that can be synchronized
// between local documents and external feedback systems
type FeedbackItem struct {
	ID          string            `json:"id"`
	ExternalID  string            `json:"external_id,omitempty"`
	Title       string            `json:"title"`
	Summary     string            `json:"summary,omitempty"`
	Description string            `json:"description"`
	Status      string            `json:"status"`
	Priority    string            `json:"priority,omitempty"`
	Type        string            `json:"type,omitempty"` // "voc" or "vos"
	FilePath    string            `json:"file_path,omitempty"`
	Tags        []string          `json:"tags,omitempty"`
	Categories  []string          `json:"categories,omitempty"` // 0..N category IDs
	Votes       int               `json:"votes,omitempty"`
	CreatedAt   string            `json:"created_at,omitempty"`
	UpdatedAt   string            `json:"updated_at,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

// ProviderConfig holds configuration for connecting to a feedback provider
type ProviderConfig struct {
	Endpoint string            `json:"endpoint"`
	APIToken string            `json:"api_token"`
	Options  map[string]string `json:"options,omitempty"`
}

// FeedbackProvider defines the interface for external feedback systems
// All synchronization logic works with this interface, enabling support
// for multiple providers (Fider, Canny, ProductBoard, etc.)
type FeedbackProvider interface {
	// Name returns the provider identifier (e.g., "fider", "canny")
	Name() string

	// Connect establishes connection to the external system
	Connect(config ProviderConfig) error

	// List retrieves all feedback items from the external system
	List() ([]FeedbackItem, error)

	// Get retrieves a specific feedback item by its external ID
	Get(id string) (*FeedbackItem, error)

	// Create adds a new feedback item to the external system
	Create(item FeedbackItem) (*FeedbackItem, error)

	// Update modifies an existing feedback item
	Update(item FeedbackItem) error

	// Delete removes a feedback item from the external system
	Delete(id string) error

	// Close releases any resources held by the provider
	Close() error
}

// ProviderRegistry manages available feedback providers
type ProviderRegistry struct {
	providers map[string]func() FeedbackProvider
}

// NewProviderRegistry creates a new provider registry
func NewProviderRegistry() *ProviderRegistry {
	return &ProviderRegistry{
		providers: make(map[string]func() FeedbackProvider),
	}
}

// Register adds a provider factory to the registry
func (r *ProviderRegistry) Register(name string, factory func() FeedbackProvider) {
	r.providers[name] = factory
}

// Get returns a new instance of the named provider
func (r *ProviderRegistry) Get(name string) (FeedbackProvider, bool) {
	factory, ok := r.providers[name]
	if !ok {
		return nil, false
	}
	return factory(), true
}

// List returns names of all registered providers
func (r *ProviderRegistry) List() []string {
	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// Global provider registry
var providerRegistry = NewProviderRegistry()

// RegisterProvider registers a provider in the global registry
func RegisterProvider(name string, factory func() FeedbackProvider) {
	providerRegistry.Register(name, factory)
}

// GetProvider returns a provider from the global registry
func GetProvider(name string) (FeedbackProvider, bool) {
	return providerRegistry.Get(name)
}

// ListProviders returns all registered provider names
func ListProviders() []string {
	return providerRegistry.List()
}

// ConflictResolution defines how to resolve sync conflicts
type ConflictResolution string

const (
	ConflictLocal     ConflictResolution = "local"     // Local version wins
	ConflictRemote    ConflictResolution = "remote"    // Remote version wins
	ConflictTimestamp ConflictResolution = "timestamp" // Newer version wins
	ConflictManual    ConflictResolution = "manual"    // Ask user
)

// SyncConflict represents a synchronization conflict
type SyncConflict struct {
	ItemID     string       `json:"item_id"`
	LocalItem  FeedbackItem `json:"local_item"`
	RemoteItem FeedbackItem `json:"remote_item"`
	Reason     string       `json:"reason"`
	Resolution string       `json:"resolution,omitempty"`
}
