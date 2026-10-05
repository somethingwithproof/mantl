package evidence

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	types "github.com/aws/aws-sdk-go-v2/service/s3/types"
	"io"
	"net/url"
	"strings"
	"time"
)

const MinimumRetentionDays = 2557

type ObjectRef struct {
	RetainUntil *time.Time `json:"retainUntil,omitempty"`
	URI         string     `json:"uri"`
	Hash        string     `json:"sha256"`
	Version     string     `json:"version"`
	Control     string     `json:"control,omitempty"`
	Resource    string     `json:"resource,omitempty"`
}
type Manifest struct {
	SchemaVersion    int         `json:"schemaVersion"`
	Audit            string      `json:"audit"`
	Run              string      `json:"run"`
	Framework        string      `json:"framework"`
	FrameworkVersion string      `json:"frameworkVersion"`
	CapturedAt       time.Time   `json:"capturedAt"`
	Objects          []ObjectRef `json:"objects"`
	Gaps             []string    `json:"coverageGaps,omitempty"`
}
type Store interface {
	Put(context.Context, string, []byte, time.Time) (ObjectRef, error)
	Get(context.Context, ObjectRef) ([]byte, error)
}
type S3API interface {
	GetBucketVersioning(context.Context, *s3.GetBucketVersioningInput, ...func(*s3.Options)) (*s3.GetBucketVersioningOutput, error)
	GetObjectLockConfiguration(context.Context, *s3.GetObjectLockConfigurationInput, ...func(*s3.Options)) (*s3.GetObjectLockConfigurationOutput, error)
	PutObject(context.Context, *s3.PutObjectInput, ...func(*s3.Options)) (*s3.PutObjectOutput, error)
	GetObject(context.Context, *s3.GetObjectInput, ...func(*s3.Options)) (*s3.GetObjectOutput, error)
}
type S3Store struct {
	Client        S3API
	Bucket        string
	KMSKey        string
	RetentionDays int
}

func Hash(data []byte) string { sum := sha256.Sum256(data); return hex.EncodeToString(sum[:]) }
func (s *S3Store) Put(ctx context.Context, key string, data []byte, at time.Time) (ObjectRef, error) {
	versioning, err := s.Client.GetBucketVersioning(ctx, &s3.GetBucketVersioningInput{Bucket: aws.String(s.Bucket)})
	if err != nil {
		return ObjectRef{}, fmt.Errorf("verify evidence versioning: %w", err)
	}
	if versioning.Status != types.BucketVersioningStatusEnabled {
		return ObjectRef{}, fmt.Errorf("evidence bucket versioning must be enabled")
	}
	lock, err := s.Client.GetObjectLockConfiguration(ctx, &s3.GetObjectLockConfigurationInput{Bucket: aws.String(s.Bucket)})
	if err != nil {
		return ObjectRef{}, fmt.Errorf("verify evidence object lock: %w", err)
	}
	if lock.ObjectLockConfiguration == nil || lock.ObjectLockConfiguration.ObjectLockEnabled != types.ObjectLockEnabledEnabled {
		return ObjectRef{}, fmt.Errorf("evidence bucket Object Lock must be enabled")
	}
	days := s.RetentionDays
	if days < MinimumRetentionDays {
		days = MinimumRetentionDays
	}
	retain := at.AddDate(0, 0, days).UTC().Truncate(time.Second).Add(time.Second)
	encryption := types.ServerSideEncryptionAes256
	if s.KMSKey != "" {
		encryption = types.ServerSideEncryptionAwsKms
	}
	hash := Hash(data)
	key = key + "/" + hash + ".json"
	in := &s3.PutObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(key), Body: strings.NewReader(string(data)), ContentType: aws.String("application/json"), ServerSideEncryption: encryption, ObjectLockMode: types.ObjectLockModeCompliance, ObjectLockRetainUntilDate: &retain, Metadata: map[string]string{"mantl-content-hash": hash}}
	if s.KMSKey != "" {
		in.SSEKMSKeyId = aws.String(s.KMSKey)
	}
	out, err := s.Client.PutObject(ctx, in)
	if err != nil {
		return ObjectRef{}, fmt.Errorf("store immutable evidence: %w", err)
	}
	if aws.ToString(out.VersionId) == "" {
		return ObjectRef{}, fmt.Errorf("evidence upload returned no object version")
	}
	ref := ObjectRef{URI: "s3://" + s.Bucket + "/" + key, Hash: hash, Version: aws.ToString(out.VersionId), RetainUntil: &retain}
	if _, err := s.Get(ctx, ref); err != nil {
		return ObjectRef{}, fmt.Errorf("verify uploaded evidence: %w", err)
	}
	return ref, nil
}
func (s *S3Store) Get(ctx context.Context, ref ObjectRef) ([]byte, error) {
	u, err := url.Parse(ref.URI)
	if err != nil || u.Scheme != "s3" || u.Host != s.Bucket || ref.Version == "" {
		return nil, fmt.Errorf("invalid versioned evidence reference")
	}
	out, err := s.Client.GetObject(ctx, &s3.GetObjectInput{Bucket: aws.String(s.Bucket), Key: aws.String(strings.TrimPrefix(u.Path, "/")), VersionId: aws.String(ref.Version)})
	if err != nil {
		return nil, fmt.Errorf("read evidence: %w", err)
	}
	defer out.Body.Close()
	if out.ObjectLockMode != types.ObjectLockModeCompliance || out.ObjectLockRetainUntilDate == nil || out.ObjectLockRetainUntilDate.Before(time.Now().UTC()) {
		return nil, fmt.Errorf("evidence object lacks active COMPLIANCE retention")
	}
	if ref.RetainUntil != nil && out.ObjectLockRetainUntilDate.Before(*ref.RetainUntil) {
		return nil, fmt.Errorf("evidence retention shorter than recorded minimum")
	}
	data, err := io.ReadAll(io.LimitReader(out.Body, 32*1024*1024+1))
	if err != nil {
		return nil, fmt.Errorf("read evidence body: %w", err)
	}
	if len(data) > 32*1024*1024 || Hash(data) != ref.Hash {
		return nil, fmt.Errorf("evidence hash mismatch or payload too large")
	}
	return data, nil
}
