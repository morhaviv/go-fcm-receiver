package ecego

type EngineOption interface {
	apply(*Engine)
}

type keyLabelOption string

func (k keyLabelOption) apply(e *Engine) { e.keyLabel = string(k) }

// WithKeyLabel sets a key label to use
func WithKeyLabel(keyLabel string) EngineOption { return keyLabelOption(keyLabel) }

type authSecretOption []byte

func (a authSecretOption) apply(e *Engine) { e.authSecret = a }

// WithAuthSecret specifies auth secret for shared key derivation
func WithAuthSecret(authSecret []byte) EngineOption { return authSecretOption(authSecret) }
