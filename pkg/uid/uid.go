package uid

type Mapper interface {
	FindOrAdd(id string) uint32
	Remove(id string)
	Validity() uint32
}
