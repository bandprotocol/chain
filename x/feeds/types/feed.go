package types

// Validate validates a Feed.
func (f *Feed) Validate() error {
	if err := validateInt64("interval", true, f.Interval); err != nil {
		return err
	}

	return nil
}
