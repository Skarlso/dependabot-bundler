package pgp

import (
	"bytes"
	"crypto"
	"fmt"

	"golang.org/x/crypto/openpgp"
	"golang.org/x/crypto/openpgp/armor"
	"golang.org/x/crypto/openpgp/packet"
)

const (
	defaultBitSize          = 4096
	defaultCompressionLevel = 9
	oneYear                 = 86400 * 365
)

// Entity defines an openpgp signer.
type Entity struct {
	Name       string
	Email      string
	BitSize    int
	PublicKey  []byte
	PrivateKey []byte
}

func (e *Entity) GetEntity() (*openpgp.Entity, error) {
	publicKeyPacket, err := e.getKeyPacket(e.PublicKey)
	if err != nil {
		return nil, err
	}

	var privKey *packet.PrivateKey

	if len(e.PrivateKey) > 0 {
		privateKeyPacket, err := e.getKeyPacket(e.PrivateKey)
		if err != nil {
			return nil, err
		}

		var ok bool

		privKey, ok = privateKeyPacket.(*packet.PrivateKey)
		if !ok {
			return nil, fmt.Errorf("private key is not of the right format")
		}
	}

	pubKey, ok := publicKeyPacket.(*packet.PublicKey)
	if !ok {
		return nil, fmt.Errorf("public key is not of the right format")
	}

	return e.createEntityFromKeys(pubKey, privKey)
}

// From https://gist.github.com/eliquious/9e96017f47d9bd43cdf9
func (e *Entity) createEntityFromKeys(pubKey *packet.PublicKey, privKey *packet.PrivateKey) (*openpgp.Entity, error) {
	config := packet.Config{
		DefaultHash:            crypto.SHA256,
		DefaultCipher:          packet.CipherAES256,
		DefaultCompressionAlgo: packet.CompressionZLIB,
		CompressionConfig: &packet.CompressionConfig{
			Level: defaultCompressionLevel,
		},
		RSABits: defaultBitSize,
	}
	currentTime := config.Now()
	uid := packet.NewUserId("", "", "")

	oe := openpgp.Entity{
		PrimaryKey: pubKey,
		PrivateKey: privKey,
		Identities: make(map[string]*openpgp.Identity),
	}
	isPrimaryID := false

	oe.Identities[uid.Id] = &openpgp.Identity{
		Name:   uid.Name,
		UserId: uid,
		SelfSignature: &packet.Signature{
			CreationTime: currentTime,
			SigType:      packet.SigTypePositiveCert,
			PubKeyAlgo:   packet.PubKeyAlgoRSA,
			Hash:         config.Hash(),
			IsPrimaryId:  &isPrimaryID,
			FlagsValid:   true,
			FlagSign:     true,
			FlagCertify:  true,
			IssuerKeyId:  &oe.PrimaryKey.KeyId,
		},
	}

	keyLifetimeSecs := uint32(oneYear)

	oe.Subkeys = make([]openpgp.Subkey, 1)
	oe.Subkeys[0] = openpgp.Subkey{
		PublicKey:  pubKey,
		PrivateKey: privKey,
		Sig: &packet.Signature{
			CreationTime:              currentTime,
			SigType:                   packet.SigTypeSubkeyBinding,
			PubKeyAlgo:                packet.PubKeyAlgoRSA,
			Hash:                      config.Hash(),
			PreferredHash:             []uint8{8}, // SHA-256
			FlagsValid:                true,
			FlagEncryptStorage:        true,
			FlagEncryptCommunications: true,
			IssuerKeyId:               &oe.PrimaryKey.KeyId,
			KeyLifetimeSecs:           &keyLifetimeSecs,
		},
	}

	return &oe, nil
}

func (e *Entity) getKeyPacket(key []byte) (packet.Packet, error) {
	keyReader := bytes.NewReader(key)

	block, err := armor.Decode(keyReader)
	if err != nil {
		return nil, err
	}

	packetReader := packet.NewReader(block.Body)

	pkt, err := packetReader.Next()
	if err != nil {
		return nil, err
	}

	return pkt, nil
}
