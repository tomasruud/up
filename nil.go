package up

import "database/sql"

// NilStore is a StateStore implementation that does nothing, useful for testing or using if you do not need a store.
type NilStore struct {
	Value *int
}

func (s *NilStore) Prepare(_ *sql.Tx) error {
	return nil
}

func (s *NilStore) Get(_ *sql.Tx) (*int, error) {
	return s.Value, nil
}

func (s *NilStore) Set(_ *sql.Tx, i int) error {
	s.Value = &i
	return nil
}
