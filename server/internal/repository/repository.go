package repository

// TODO: move to another package.

type Repository interface {
	Close() error
	AddSubscriber(int64) error
	RemoveSubscriber(int64) error
	Subscribers() ([]int64, error)

	ManageDebug(enable bool) error
	DebugEnabled() (bool, error)
}
