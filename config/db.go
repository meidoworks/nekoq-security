package config

import (
	"container/list"
	"errors"
	"fmt"

	"go.etcd.io/bbolt"
)

func (c *NekoQSecurityContainer) ListAllProviderBuckets() ([]string, error) {

	li := list.New()
	err := c.db.View(func(tx *bbolt.Tx) error {
		for _, v := range moduleNamespace {
			b := tx.Bucket([]byte(v.Namespace))
			if b != nil {
				li.PushBack(v)
			}
		}
		return nil
	})

	if err != nil {
		return nil, err
	}

	r := []string{}
	for e := li.Front(); e != nil; e = e.Next() {
		r = append(r, e.Value.(string))
	}
	return r, nil
}

func (c *NekoQSecurityContainer) DoTxWithinBucket(bucket string, fn func(*bbolt.Bucket) error) error {
	err := c.db.Update(func(tx *bbolt.Tx) error {
		b := tx.Bucket([]byte(bucket))
		if b == nil {
			return errors.New(fmt.Sprint("no bucket:", bucket, " found"))
		}
		return fn(b)
	})
	return err
}
