package uid

import (
	"encoding/gob"
	"os"
	"sync"
)

type UidMapper struct {
	Validity uint32
	Next     uint32
	Values   map[string]uint32
	Path     string
	lock     *sync.RWMutex
}

func New(path string) (*UidMapper, error) {
	mapper := UidMapper{}

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

func (u *UidMapper) Flush() error {
	u.lock.Lock()
	defer u.lock.Unlock()

	file, err := os.OpenFile(u.Path, os.O_WRONLY, 644)
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

func (u *UidMapper) FindOrAdd(id string) uint32 {
	u.lock.Lock()
	defer u.lock.Unlock()

	if _, ok := u.Values[id]; !ok {
		u.Values[id] = u.Next
		u.Next++
	}

	return u.Values[id]
}

func (u *UidMapper) Remove(id string) {
	delete(u.Values, id)
}

func (u *UidMapper) setDefaults(path string) *UidMapper {
	if u.Validity == 0 {
		u.Validity = 1
	}

	if u.Next == 0 {
		u.Next = 1
	}

	if u.Values == nil {
		u.Values = make(map[string]uint32)
	}

	if u.Path == "" {
		u.Path = path
	}
	return u
}
