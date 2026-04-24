package storage

type Store interface {
	Put(collection, id string, value any) error
	Get(collection, id string, dest any) error
	List(collection string, dest any) error
}
