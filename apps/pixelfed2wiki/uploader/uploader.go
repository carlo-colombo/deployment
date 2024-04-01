package uploader

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"hash"
	"image"
	"image/jpeg"
	"image/png"
	"io"
	"net/http"

	"github.com/minio/minio-go/v7"
	"github.com/nfnt/resize"
)

type Uploader struct {
	mc         *minio.Client
	bucket     string
	hasher     hash.Hash
	imageBytes []byte
	imageName  string
	mimeType   string
	Err        error
}

func NewUploader(bucket string, mc *minio.Client, imageBytes []byte) Uploader {
	r := bytes.NewReader(imageBytes)
	hasher := sha256.New()

	if _, err := io.Copy(hasher, r); err != nil {
		return Uploader{
			Err: fmt.Errorf("failed to calculate hash: %w", err),
		}
	}

	mimeType := http.DetectContentType(imageBytes)

	return Uploader{
		bucket:     bucket,
		mc:         mc,
		imageBytes: imageBytes,
		imageName:  fmt.Sprintf("%x", hasher.Sum(nil)),
		mimeType:   mimeType,
	}
}

func (u *Uploader) UploadImage() (string, error) {
	if u.Err != nil {
		return "", u.Err
	}

	info, err := u.mc.PutObject(context.TODO(), u.bucket, "images/"+u.imageName,
		bytes.NewReader(u.imageBytes), int64(len(u.imageBytes)), minio.PutObjectOptions{
			CacheControl: "max-age=604800, must-revalidate",
		})

	if err != nil {
		return "", fmt.Errorf("cannot upload image with key '%s': %w", u.imageName, err)
	}
	return fmt.Sprintf("%s/%s", u.getPublicBucketUrl(), info.Key), nil
}

func (u Uploader) getPublicBucketUrl() string {
	bucketUrl := *u.mc.EndpointURL()
	bucketUrl.Host = u.bucket + "." + bucketUrl.Host
	return bucketUrl.String()
}

func (u *Uploader) UploadThumbnail() (string, error) {
	if u.Err != nil {
		return "", u.Err
	}

	var img image.Image
	var err error

	switch u.mimeType {
	case "image/png":
		img, err = png.Decode(bytes.NewReader(u.imageBytes))
	case "image/jpeg":
		img, err = jpeg.Decode(bytes.NewReader(u.imageBytes))
	default:
		err = fmt.Errorf("mimetype not recognized %s", u.mimeType)
	}

	if err != nil {
		return "", fmt.Errorf("failed decoding image: %w", err)
	}

	thumb := resize.Thumbnail(360, 360, img, resize.Lanczos2)

	buf := bytes.NewBuffer([]byte{})

	err = jpeg.Encode(buf, thumb, nil)

	if err != nil {
		return "", fmt.Errorf("failed encoding thumb: %w", err)
	}

	info, err := u.mc.PutObject(context.TODO(), u.bucket, "thumbs/"+u.imageName,
		buf, int64(buf.Len()), minio.PutObjectOptions{
			CacheControl: "max-age=604800, must-revalidate",
		})

	if err != nil {
		return "", fmt.Errorf("cannot upload thumbnail: %w", err)
	}

	return fmt.Sprintf("%s/%s", u.getPublicBucketUrl(), info.Key), nil
}
