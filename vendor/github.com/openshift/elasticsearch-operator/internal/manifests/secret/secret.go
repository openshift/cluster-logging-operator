package secret

import (
	"context"
	"crypto/sha256"
	"fmt"
	"sort"

	"github.com/ViaQ/logerr/kverrors"
	"github.com/ViaQ/logerr/log"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/equality"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/client-go/util/retry"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

// EqualityFunc is the type for functions that compare two secrets.
// Return true if two secrets are equal.
type EqualityFunc func(current, desired *corev1.Secret) bool

// MutateFunc is the type for functions that mutate the current secret
// by applying the values from the desired secret.
type MutateFunc func(current, desired *corev1.Secret)

// Get returns the k8s secret for the given object key or an error.
func Get(ctx context.Context, c client.Client, key client.ObjectKey) (*corev1.Secret, error) {
	s := New(key.Name, key.Namespace, map[string][]byte{})

	if err := c.Get(ctx, key, s); err != nil {
		return s, kverrors.Wrap(err, "failed to get secret",
			"name", s.Name,
			"namespace", s.Namespace,
		)
	}

	return s, nil
}

// GetDataSHA256 returns the sha256 checksum of the secret data keys
func GetDataSHA256(ctx context.Context, c client.Client, key client.ObjectKey) string {
	hash := ""

	sec, err := Get(ctx, c, key)
	if err != nil {
		return hash
	}

	dataHashes := make(map[string][32]byte)

	for key, data := range sec.Data {
		dataHashes[key] = sha256.Sum256([]byte(data))
	}

	sortedKeys := []string{}
	for key := range dataHashes {
		sortedKeys = append(sortedKeys, key)
	}

	sort.Strings(sortedKeys)

	for _, key := range sortedKeys {
		hash = fmt.Sprintf("%s%s", hash, dataHashes[key])
	}

	return hash
}

// CreateOrUpdate attempts first to create the given secret. If the
// secret already exists and the provided comparison func detects any changes
// an update is attempted. Updates are retried with backoff (See retry.DefaultRetry).
// Returns on failure an non-nil error.
func CreateOrUpdate(ctx context.Context, c client.Client, s *corev1.Secret, equal EqualityFunc, mutate MutateFunc) error {
	err := c.Create(ctx, s)
	if err == nil {
		return nil
	}

	if !apierrors.IsAlreadyExists(kverrors.Root(err)) {
		return kverrors.Wrap(err, "failed to create secret",
			"name", s.Name,
			"namespace", s.Namespace,
		)
	}

	current := &corev1.Secret{}
	key := client.ObjectKey{Name: s.Name, Namespace: s.Namespace}
	err = c.Get(ctx, key, current)
	if err != nil {
		return kverrors.Wrap(err, "failed to get secret",
			"name", s.Name,
			"namespace", s.Namespace,
		)
	}

	if !equal(current, s) {
		err := retry.RetryOnConflict(retry.DefaultRetry, func() error {
			if err := c.Get(ctx, key, current); err != nil {
				log.Error(err, "failed to get secret", s.Name)
				return err
			}

			mutate(current, s)
			if err := c.Update(ctx, current); err != nil {
				log.Error(err, "failed to update secret", s.Name)
				return err
			}
			return nil
		})
		if err != nil {
			return kverrors.Wrap(err, "failed to update secret",
				"name", s.Name,
				"namespace", s.Namespace,
			)
		}
		return nil
	}

	return nil
}

// DataEqual returns true only if the data of current and desird are exactly same.
func DataEqual(current, desired *corev1.Secret) bool {
	return equality.Semantic.DeepEqual(current.Data, desired.Data)
}

// MutateDataOnly is a default mutation function for services
// that copies only the data field from desired to current.
func MutateDataOnly(current, desired *corev1.Secret) {
	current.Data = desired.Data
}
