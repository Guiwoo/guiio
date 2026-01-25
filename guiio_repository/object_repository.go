package guiio_repository

import (
	"context"
	"fmt"

	"guiio/ent"
	"guiio/ent/object"
	"guiio/ent/objectmetadata"
)

type ObjectRepository interface {
	UpsertObject(ctx context.Context, in ObjectUpsertInput) (*ent.Object, error)
	GetObject(ctx context.Context, bucketName, objectName string) (*ent.Object, error)
	DeleteObject(ctx context.Context, bucketName, objectName string) error
}

type ObjectUpsertInput struct {
	BucketName  string
	ObjectName  string
	StoragePath string
	ContentType string
	Size        int64
	ETag        string
	Metadata    map[string]string
}

type objectRepository struct {
	db *ent.Client
}

func NewObjectRepository(db *ent.Client) ObjectRepository {
	return &objectRepository{db: db}
}

func (r *objectRepository) UpsertObject(ctx context.Context, in ObjectUpsertInput) (*ent.Object, error) {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return nil, err
	}

	obj, err := tx.Object.
		Query().
		Where(
			object.BucketNameEQ(in.BucketName),
			object.ObjectNameEQ(in.ObjectName),
		).
		Only(ctx)

	create := false
	if err != nil {
		if !ent.IsNotFound(err) {
			tx.Rollback()
			return nil, err
		}
		create = true
	}

	if create {
		obj, err = tx.Object.
			Create().
			SetBucketName(in.BucketName).
			SetObjectName(in.ObjectName).
			SetStoragePath(in.StoragePath).
			SetContentType(in.ContentType).
			SetSize(in.Size).
			SetEtag(in.ETag).
			Save(ctx)
	} else {
		obj, err = obj.Update().
			SetStoragePath(in.StoragePath).
			SetContentType(in.ContentType).
			SetSize(in.Size).
			SetEtag(in.ETag).
			Save(ctx)
	}

	if err != nil {
		tx.Rollback()
		return nil, err
	}

	if in.Metadata != nil {
		if err := r.replaceMetadata(ctx, tx, obj.ID, in.Metadata); err != nil {
			tx.Rollback()
			return nil, err
		}
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return obj, nil
}

func (r *objectRepository) replaceMetadata(ctx context.Context, tx *ent.Tx, objectID int, meta map[string]string) error {
	if _, err := tx.ObjectMetadata.
		Delete().
		Where(objectmetadata.ObjectIDEQ(objectID)).
		Exec(ctx); err != nil {
		return fmt.Errorf("clear metadata: %w", err)
	}

	bulk := make([]*ent.ObjectMetadataCreate, 0, len(meta))
	for k, v := range meta {
		bulk = append(bulk, tx.ObjectMetadata.
			Create().
			SetObjectID(objectID).
			SetKey(k).
			SetValue(v))
	}

	if len(bulk) == 0 {
		return nil
	}

	if err := tx.ObjectMetadata.CreateBulk(bulk...).Exec(ctx); err != nil {
		return fmt.Errorf("create metadata: %w", err)
	}

	return nil
}

func (r *objectRepository) GetObject(ctx context.Context, bucketName, objectName string) (*ent.Object, error) {
	return r.db.Object.
		Query().
		Where(
			object.BucketNameEQ(bucketName),
			object.ObjectNameEQ(objectName),
		).
		WithMetadata().
		Only(ctx)
}

func (r *objectRepository) DeleteObject(ctx context.Context, bucketName, objectName string) error {
	tx, err := r.db.Tx(ctx)
	if err != nil {
		return err
	}

	obj, err := tx.Object.
		Query().
		Where(
			object.BucketNameEQ(bucketName),
			object.ObjectNameEQ(objectName),
		).
		Only(ctx)
	if err != nil {
		tx.Rollback()
		return err
	}

	if _, err := tx.ObjectMetadata.Delete().Where(objectmetadata.ObjectIDEQ(obj.ID)).Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete metadata: %w", err)
	}

	if err := tx.Object.DeleteOneID(obj.ID).Exec(ctx); err != nil {
		tx.Rollback()
		return fmt.Errorf("delete object: %w", err)
	}

	return tx.Commit()
}
