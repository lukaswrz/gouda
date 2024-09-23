// Package upsched provides utilities for managing and scheduling multi-part
// uploads with a timeout mechanism. It allows appending chunks of data to
// a destination and automatically finalizes the upload if a specified
// timeout is reached.
package upsched

import (
	"errors"
	"fmt"
	"io"
	"mime/multipart"
	"os"
	"time"

	"github.com/alphadose/haxmap"
	"golang.org/x/exp/constraints"
)

const (
	// AppendOpenFlags is the recommended flag set for opening a file to
	// which chunks will be appended during the upload process.
	AppendOpenFlags = os.O_APPEND | os.O_CREATE | os.O_WRONLY
)

// upload holds the state for a single upload, including its timeout
// duration and an associated timer.
type upload struct {
	timeout time.Duration
	timer   *time.Timer
}

// UploadKey defines the set of types that can be used as keys in the
// UploadScheduler. It can be any integer type or a string.
type UploadKey interface {
	constraints.Integer | ~string
}

// UploadScheduler manages the scheduling of multi-part uploads. It maintains
// a map of active uploads and handles the appending of chunks, as well as
// the automatic finalization of uploads based on a timeout.
type UploadScheduler[K UploadKey] struct {
	m *haxmap.Map[K, upload]
}

// NewScheduler creates a new UploadScheduler instance. It returns an
// UploadScheduler configured to manage uploads keyed by the specified type.
func NewScheduler[K UploadKey]() UploadScheduler[K] {
	return UploadScheduler[K]{
		m: haxmap.New[K, upload](),
	}
}

// Prepare initializes an upload with the given key and timeout duration.
// If an upload with the specified key already exists, an error is returned.
// The provided callback function is called with the key and an error if the
// upload times out before it is finished. This function should be called
// before any chunks are appended.
//
// If the upload is successfully initialized, a timer is started based on the
// provided timeout duration. If the timer expires before the upload is
// finished, the callback function is invoked.
//
// Returns an error if the key already exists in the scheduler.
func (us UploadScheduler[K]) Prepare(k K, timeout time.Duration, cb func(K, error)) error {
	_, ok := us.m.Get(k)
	if ok {
		return errors.New("upload key already exists")
	}

	us.m.Set(
		k,
		upload{
			timeout: timeout,
			timer: time.AfterFunc(
				time.Second*time.Duration(timeout),
				func() {
					if _, ok := us.m.Get(k); ok {
						err := us.Finish(k)
						cb(k, err)
					}
				},
			),
		},
	)

	return nil
}

// Append appends a chunk of data to the destination writer associated with
// the given key. It resets the upload's timer to the initial timeout duration
// upon a successful append. If the key does not exist, an error is returned.
//
// It is recommended to use AppendOpenFlags for actual files that are passed
// to this function.
func (us UploadScheduler[K]) Append(k K, chunk multipart.File, dst io.Writer) error {
	u, ok := us.m.Get(k)
	if !ok {
		return errors.New("upload key does not exist")
	}

	u.timer.Stop()
	defer u.timer.Reset(u.timeout)

	_, err := io.Copy(dst, chunk)
	if err != nil {
		return fmt.Errorf("unable to append chunk to destination file: %w", err)
	}

	return nil
}

// Finish finalizes the upload associated with the given key. It stops the
// associated timer and removes the upload from the scheduler's internal map.
// If the key does not exist, an error is returned.
func (us UploadScheduler[K]) Finish(k K) error {
	u, ok := us.m.Get(k)
	if !ok {
		return errors.New("upload key does not exist")
	}

	u.timer.Stop()
	us.m.Del(k)

	return nil
}