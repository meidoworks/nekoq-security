package config

import (
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"log"
	"time"

	"goimport.moetang.info/nekoq-security/alg/aesutils"
	"goimport.moetang.info/nekoq-security/alg/shamir"

	"go.etcd.io/bbolt"
)

const (
	MaxShares = 5
	MinShares = 3
)

type NekoQSecurityConfig struct {
	NekoQSecurity struct {
		MasterKey struct {
			Type string `toml:"type"`
		} `toml:"masterkey"`
		Storage struct {
			Path string `toml:"path"`
		} `toml:"storage"`
	} `toml:"nekoq-security"`

	container *NekoQSecurityContainer
}

type NekoQSecurityContainer struct {
	db *bbolt.DB

	MasterUnlock bool

	ShamirShards []string
	MasterKey    []byte
}

func (c *NekoQSecurityConfig) Validate() error {
	switch c.NekoQSecurity.MasterKey.Type {
	case "shamir":
	default:
		return errors.New("unknown master key type")
	}
	if len(c.NekoQSecurity.Storage.Path) == 0 {
		return errors.New("no path for storage")
	}
	return nil
}

func (c *NekoQSecurityConfig) Init() error {
	c.container = new(NekoQSecurityContainer)

	db, err := bbolt.Open(c.NekoQSecurity.Storage.Path, 0666, &bbolt.Options{
		Timeout: 5 * time.Second,
	})
	if err != nil {
		return err
	}
	c.container.db = db

	return nil
}

func (c *NekoQSecurityConfig) IsMasterUnlock() bool {
	return c.container.MasterUnlock
}

func (c *NekoQSecurityConfig) FeedShamirKey(key string) bool {
	if c.IsMasterUnlock() {
		return true
	}
	if len(c.container.ShamirShards) >= MaxShares {
		return false
	}

	c.container.ShamirShards = append(c.container.ShamirShards, key)
	if len(c.container.ShamirShards) < MinShares {
		return false
	}

	m, err := shamir.CombineShamirString(c.container.ShamirShards)
	if err != nil {
		log.Println("[ERROR] combine shamir key error.", err)
		return false
	}

	err = c.container.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte("global"))
		if b == nil {
			bucket, err := tx.CreateBucket([]byte("global"))
			if err != nil {
				return err
			}
			b = bucket
		}
		v := b.Get([]byte("nekoq-security.init"))
		if len(v) == 0 {
			// need to init
			rb := make([]byte, 8)
			_, err := rand.Read(rb)
			if err != nil {
				return err
			}

			rb = append([]byte(hex.EncodeToString(rb)), []byte("_nekoq-security")...)

			enc, err := aesutils.Encrypt(rb, m)
			if err != nil {
				return err
			}
			err = b.Put([]byte("nekoq-security.init"), enc)
			if err != nil {
				return err
			}
		} else {
			// check master key
			dec, err := aesutils.Decrypt(v, m)
			if err != nil {
				return err
			}
			if !bytes.HasSuffix(dec, []byte("_nekoq-security")) {
				return errors.New("masterkey cannot decrypt init value")
			}
			// decrypt success and init masterkey
			c.container.MasterUnlock = true
			c.container.MasterKey = m
		}
		return nil
	})
	if err != nil {
		log.Println("[ERROR] FeedShamirKey error.", err)
		return false
	}
	return true
}

func (c *NekoQSecurityConfig) ResetMasterKeyWhileUnlocking() {
	if c.container.MasterUnlock {
		return
	}
	c.container.ShamirShards = nil
}
