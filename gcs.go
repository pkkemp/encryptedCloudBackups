package main

import (
	"cloud.google.com/go/storage"
	"context"
	"fmt"
	"google.golang.org/api/option"
	"io"
	"log"
	"os"
	"time"
)

// uploadEncryptedFile writes an object using AES-256 encryption key.
//func uploadEncryptedFile(w io.Writer, bucket, object string, secretKey []byte) error {
//	// bucket := "bucket-name"
//	// object := "object-name"
//	// secretKey := []byte("secret-key")
//	ctx := context.Background()
//	client, err := storage.NewClient(ctx, option.WithCredentialsFile(SERVICE_ACCOUNT_KEY_PATH))
//	if err != nil {
//		return fmt.Errorf("storage.NewClient: %w", err)
//	}
//	defer client.Close()
//
//	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
//	defer cancel()
//
//	o := client.Bucket(bucket).Object(object)
//
//	// Optional: set a generation-match precondition to avoid potential race
//	// conditions and data corruptions. The request to upload is aborted if the
//	// object's generation number does not match your precondition.
//	// For an object that does not yet exist, set the DoesNotExist precondition.
//	o = o.If(storage.Conditions{DoesNotExist: true})
//	// If the live object already exists in your bucket, set instead a
//	// generation-match precondition using the live object's generation number.
//	// attrs, err := o.Attrs(ctx)
//	// if err != nil {
//	//      return fmt.Errorf("object.Attrs: %w", err)
//	// }
//	// o = o.If(storage.Conditions{GenerationMatch: attrs.Generation})
//
//	// Encrypt the object's contents.
//	wc := o.Key(secretKey).NewWriter(ctx)
//	if _, err := wc.Write([]byte("top secret")); err != nil {
//		return fmt.Errorf("Writer.Write: %w", err)
//	}
//	if err := wc.Close(); err != nil {
//		return fmt.Errorf("Writer.Close: %w", err)
//	}
//	fmt.Fprintf(w, "Uploaded encrypted object %v.\n", object)
//	return nil
//}

// uploadEncryptedFile writes an object using AES-256 encryption key.
func uploadEncryptedFile(bucket, object string, secretKey []byte, filePath string) (bool, error) {
	// bucket := "bucket-name"
	// object := "object-name"
	// secretKey := []byte("secret-key")
	ctx := context.Background()
	client, err := storage.NewClient(ctx, option.WithCredentialsFile(SERVICE_ACCOUNT_KEY_PATH))
	if err != nil {
		return false, fmt.Errorf("storage.NewClient: %w", err)
	}
	defer client.Close()

	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	o := client.Bucket(bucket).Object(object)

	// Optional: set a generation-match precondition to avoid potential race
	// conditions and data corruptions. The request to upload is aborted if the
	// object's generation number does not match your precondition.
	// For an object that does not yet exist, set the DoesNotExist precondition.
	o = o.If(storage.Conditions{DoesNotExist: true})

	// Open the file for reading
	file, err := os.Open(filePath)
	if err != nil {
		return false, fmt.Errorf("os.Open: %w", err)
	}
	defer file.Close()

	// Encrypt the object's contents.
	wc := o.Key(secretKey).NewWriter(ctx)
	if _, err := io.Copy(wc, file); err != nil {
		return false, fmt.Errorf("io.Copy: %w", err)
	}
	if err := wc.Close(); err != nil {
		return false, fmt.Errorf("Writer.Close: %w", err)
	}
	log.Printf("Uploaded encrypted object %v.\n", object)
	return true, nil
}
