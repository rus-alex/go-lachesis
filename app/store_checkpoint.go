package app

// SetCheckpoint save Checkpoint.
// Checkpoint is seldom read; so no cache.
func (s *Store) SetCheckpoint(cp Checkpoint) {
	// TODO: implement it
	/*
		const key = "c"
		cp.RWMutex = nil
		s.set(s.table.Checkpoint, []byte(key), cp)
	*/
}

// GetCheckpoint returns stored Checkpoint.
// State is seldom read; so no cache.
func (s *Store) GetCheckpoint() *Checkpoint {
	// TODO: implement it
	/*
		const key = "c"
		cp, _ := s.get(s.table.Checkpoint, []byte(key), &Checkpoint{}).(*Checkpoint)
		cp.RWMutex = new(sync.RWMutex)
		return cp
	*/
	return nil
}
