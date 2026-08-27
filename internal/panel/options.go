package panel

// Options controls optional dashboard rendering behavior.
type Options struct {
	// ShowSources renders collapsible per-section raw request/response traces (dev only).
	ShowSources bool
	// PublicEVM is the HTTPS JSON-RPC URL injected into the Delegate page for MetaMask.
	PublicEVM string
}
