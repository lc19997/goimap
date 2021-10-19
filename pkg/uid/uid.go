package uid

import (
	"encoding/gob"
	"os"
)

type UidMapper struct {
	Validity uint32
	Next     uint32

	values map[string]uint32
	path   string
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
		if err := encoder.Encode(mapper.setDefaults()); err != nil {
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

	return mapper.setDefaults(), nil
}

func (u *UidMapper) Flush() error {
	file, err := os.Open(u.path)
	if err != nil {
		return err
	}
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(*u); err != nil {
		return err
	}

	return nil
}

func (u *UidMapper) FindOrAdd(id string) uint32 {
	if _, ok := u.values[id]; !ok {
		u.values[id] = u.Next
		u.Next++
	}

	return u.values[id]
}

func (u *UidMapper) setDefaults() *UidMapper {
	if u.Validity == 0 {
		u.Validity = 1
	}

	if u.Next == 0 {
		u.Next = 1
	}

	if u.values == nil {
		u.values = make(map[string]uint32)
	}

	return u
}