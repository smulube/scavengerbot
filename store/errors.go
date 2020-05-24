package store

// Error is a string type alias used to create const errors
type Error string

// Error is our implementation of the error interface
func (e Error) Error() string {
	return string(e)
}

const (
	// ErrNotFound is sentinel error returned when a resource is not present in
	// the specified bucket
	ErrNotFound = Error("Resource not found")

	// ErrBucketNotFound is error returned when a bucket does not exist
	ErrBucketNotFound = Error("Bucket not found")
)
