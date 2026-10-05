package evidence

import (
	"context"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
	"strings"
	"testing"
	"time"
)

type fakeS3 struct {
	versioning types.BucketVersioningStatus
	lock       types.ObjectLockEnabled
	input      *s3.PutObjectInput
	data       string
}

func (f *fakeS3) GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput, ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error) {
	return &s3.GetBucketVersioningOutput{Status: f.versioning}, nil
}
func (f *fakeS3) GetObjectLockConfiguration(context.Context, *s3.GetObjectLockConfigurationInput, ...func(*s3.Options)) (*s3.GetObjectLockConfigurationOutput, error) {
	return &s3.GetObjectLockConfigurationOutput{ObjectLockConfiguration: &types.ObjectLockConfiguration{ObjectLockEnabled: f.lock}}, nil
}
func (f *fakeS3) PutObject(_ context.Context, in *s3.PutObjectInput, _ ...func(*s3.Options)) (*s3.PutObjectOutput, error) {
	f.input = in
	bytes, _ := io.ReadAll(in.Body)
	f.data = string(bytes)
	return &s3.PutObjectOutput{VersionId: aws.String("version")}, nil
}
func (f *fakeS3) GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error) {
	return &s3.GetObjectOutput{Body: io.NopCloser(strings.NewReader(f.data)), ObjectLockMode: types.ObjectLockModeCompliance, ObjectLockRetainUntilDate: f.input.ObjectLockRetainUntilDate}, nil
}
func TestImmutableStoreFailsClosed(t *testing.T) {
	api := &fakeS3{}
	store := S3Store{Client: api, Bucket: "evidence", RetentionDays: 365}
	at := time.Now()
	if _, err := store.Put(context.Background(), "run", []byte("payload"), at); err == nil {
		t.Fatal("unversioned bucket accepted")
	}
	api.versioning = types.BucketVersioningStatusEnabled
	if _, err := store.Put(context.Background(), "run", []byte("payload"), at); err == nil {
		t.Fatal("unlocked bucket accepted")
	}
	api.lock = types.ObjectLockEnabledEnabled
	ref, err := store.Put(context.Background(), "run", []byte("payload"), at)
	if err != nil {
		t.Fatal(err)
	}
	if ref.Version == "" || api.input.ObjectLockMode != types.ObjectLockModeCompliance || api.input.ObjectLockRetainUntilDate.Before(at.AddDate(0, 0, MinimumRetentionDays)) {
		t.Fatal("immutable retention contract not enforced")
	}
	api.data = "tampered"
	if _, err := store.Get(context.Background(), ref); err == nil {
		t.Fatal("tampered evidence accepted")
	}
	api.data = "payload"
	if _, err := store.Get(context.Background(), ref); err != nil {
		t.Fatal(err)
	}
}
