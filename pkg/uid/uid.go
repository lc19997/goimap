// Package uid contains a simple mapper for UID (an integer value) and the notmuch
// database ID (a string) and persists it to disk. UID and Validity together must
// permanently refer to a specific message, otherwise your MUA may be get confused.
package uid

import (
	"encoding/gob"
	"os"
	"sync"
)

type Mapper struct {
	Validity uint32
	Next     uint32
	Values   map[string]uint32

	path string
	lock *sync.RWMutex
}

func New(path string) (*Mapper, error) {
	mapper := Mapper{}

	_, err := os.Stat(path)
	if os.IsNotExist(err) {
		file, err := os.Create(path)
		if err != nil {
			return nil, err
		}
		encoder := gob.NewEncoder(file)
		if err := encoder.Encode(mapper.setDefaults(path)); err != nil {
			return nil, err
		}
		file.Close()
	} else if err != nil {
		return nil, err
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	err = decoder.Decode(&mapper)
	if err != nil {
		return nil, err
	}

	mapper.lock = &sync.RWMutex{}
	return mapper.setDefaults(path), nil
}

func (u *Mapper) Flush() error {
	u.lock.Lock()
	defer u.lock.Unlock()

	file, err := os.OpenFile(u.path, os.O_WRONLY, 0644)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(&u); err != nil {
		return err
	}

	return nil
}

func (u *Mapper) FindOrAdd(id string) uint32 {
	u.lock.Lock()
	defer u.lock.Unlock()

	if _, ok := u.Values[id]; !ok {
		u.Values[id] = u.Next
		u.Next++
	}

	return u.Values[id]
}

func (u *Mapper) Remove(id string) {
	u.lock.Lock()
	defer u.lock.Unlock()
	delete(u.Values, id)
}

func (u *Mapper) setDefaults(path string) *Mapper {
	if u.Validity == 0 {
		u.Validity = 1
	}

	if u.Next == 0 {
		u.Next = 1
	}

	if u.Values == nil {
		u.Values = make(map[string]uint32)
	}

	if u.path == "" {
		u.path = path
	}
	return u
}
