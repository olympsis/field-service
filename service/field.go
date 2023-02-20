package field

import (
	"context"
	"errors"
)

// Insert new user into database
func (f *FieldService) InsertField(ctx context.Context, field *Field) error {
	pong := f.PingDatabase()
	if !pong {
		return errors.New("failed to connect to database")
	}
	f.Collection.InsertOne(ctx, field)
	return nil
}

// Get user from database
func (f *FieldService) FindField(ctx context.Context, filter interface{}, field *Field) error {
	pong := f.PingDatabase()
	if !pong {
		return errors.New("failed to connect to database")
	}
	f.Collection.FindOne(ctx, filter).Decode(&field)
	return nil
}

// get users from database
func (f *FieldService) FindFields(ctx context.Context, filter interface{}, fields *[]Field) error {
	pong := f.PingDatabase()
	if !pong {
		return errors.New("failed to connect to database")
	}

	cursor, err := f.Collection.Find(ctx, filter)
	if err != nil {
		return err
	}

	for cursor.Next(context.TODO()) {
		var user Field
		err := cursor.Decode(&user)
		if err != nil {
			return err
		}
		*fields = append(*fields, user)
	}
	return nil
}

// update user in database
func (f *FieldService) UpdateField(ctx context.Context, filter interface{}, update interface{}, field *Field) error {
	pong := f.PingDatabase()
	if !pong {
		return errors.New("failed to connect to database")
	}

	// update user
	_, err := f.Collection.UpdateOne(ctx, filter, update)
	if err != nil {
		return err
	}

	// find and return updated user
	err = f.FindField(ctx, filter, field)
	if err != nil {
		return err
	}

	return nil
}

// update users in database
func (f *FieldService) UpdateFields(ctx context.Context, filter interface{}, update interface{}, fields *[]Field) error {
	pong := f.PingDatabase()
	if !pong {
		return errors.New("failed to connect to database")
	}

	// update users
	_, err := f.Collection.UpdateMany(ctx, filter, update)
	if err != nil {
		return err
	}

	// find updated users
	cursor, err := f.Collection.Find(ctx, filter)
	if err != nil {
		return err
	}

	for cursor.Next(context.TODO()) {
		var user Field
		err := cursor.Decode(&user)
		if err != nil {
			return err
		}
		*fields = append(*fields, user)
	}

	return nil
}

// delete user in database
func (f *FieldService) DeleteField(ctx context.Context, filter interface{}) error {
	pong := f.PingDatabase()
	if !pong {
		return errors.New("failed to connect to database")
	}

	// delete user
	_, err := f.Collection.DeleteOne(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}

// delete users in database
func (f *FieldService) DeleteFields(ctx context.Context, filter interface{}) error {
	pong := f.PingDatabase()
	if !pong {
		return errors.New("failed to connect to database")
	}

	// delete users
	_, err := f.Collection.DeleteMany(ctx, filter)
	if err != nil {
		return err
	}
	return nil
}
