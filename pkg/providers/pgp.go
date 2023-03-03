package providers

import "golang.org/x/crypto/openpgp"

// Entity defines the ability to define an entity.
type Entity interface {
	GetEntity() (*openpgp.Entity, error)
}
