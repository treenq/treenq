package cdk

import (
	"context"
	"fmt"

	tqsdk "github.com/treenq/treenq/pkg/sdk"
	"github.com/treenq/treenq/src/domain"
)

type Locker interface {
	Lock() error
	Unlock()
}

type Provider interface {
	CreateAppResource(ctx context.Context, image domain.Image, app tqsdk.Space) error
}

type Store interface {
	OpenResourceRecord(ctx context.Context, res tqsdk.Resource) (string, error)
	MarkResourceAsDone(ctx context.Context, id string) error
	MarkResourceAsReverted(ctx context.Context, id string) error
}

type SavedResource struct {
	tqsdk.Resource
	ID string
}

type Kit struct {
	locker   Locker
	provider Provider
	store    Store
}

func NewKit(
	locker Locker,
	provider Provider,
	store Store,
) *Kit {
	return &Kit{
		locker:   locker,
		provider: provider,
		store:    store,
	}
}

func (k *Kit) q(ctx context.Context, space tqsdk.Space, image domain.Image) error {
	if err := k.locker.Lock(); err != nil {
		return fmt.Errorf("cdk.q: failed to lock: %w", err)
	}
	defer k.locker.Unlock()

	res := space.ToResource()
	resID, err := k.store.OpenResourceRecord(ctx, res)
	if err != nil {
		return fmt.Errorf("cdk.q: failed to open resource record: %w", err)
	}

	if err := k.provider.CreateAppResource(ctx, image, space); err != nil {
		revertErr := k.store.MarkResourceAsReverted(ctx, resID)
		if revertErr != nil {
			return fmt.Errorf("cdk.q: failed to mark resource as reverted: %w", revertErr)
		}
		return fmt.Errorf("cdk.q: failed to create app resource: %w", err)
	}

	if err := k.store.MarkResourceAsDone(ctx, resID); err != nil {
		return fmt.Errorf("cdk.q: failed to mark resource as done: %w", err)
	}
	return nil
}
