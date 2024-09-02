package main

import (
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"fmt"
	"io"
	"log"
	"os"
	"path/filepath"

	_ "github.com/mattn/go-sqlite3"
)

// FileInfo represents the structure of a file record from the database
type FileInfo struct {
	ID       int
	Path     string
	Name     string
	Size     int64
	SHA256   string
	AESKey   string
	Uploaded int
}

func main() {
	// Specify the directory to traverse
	dirToTraverse := "./files"

	// Initialize the SQLite database
	db, err := sql.Open("sqlite3", "./db/files.db")
	if err != nil {
		log.Fatal(err)
	}
	defer db.Close()

	// Create table for storing file information
	createTableQuery := `
	CREATE TABLE IF NOT EXISTS files (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		path TEXT NOT NULL,
		name TEXT NOT NULL,
		size INTEGER NOT NULL,
		sha256 TEXT NOT NULL,
		aes_key BLOB NOT NULL,
		uploaded INTEGER NOT NULL DEFAULT 0
	);
	`
	_, err = db.Exec(createTableQuery)
	if err != nil {
		log.Fatal(err)
	}

	// Traverse the directory and store file info
	err = filepath.Walk(dirToTraverse, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		if !info.IsDir() {
			sha256Hash, err := hashFileSHA256(path)
			if err != nil {
				log.Println("Error hashing file:", err)
				return nil
			}

			// Check if the file hash already exists in the database
			exists, err := hashExistsInDB(db, sha256Hash)
			if err != nil {
				log.Println("Error checking hash in database:", err)
				return nil
			}

			if exists {
				log.Printf("File %s with hash %s already exists in the database, skipping.\n", path, sha256Hash)
				return nil
			}

			aesKey, err := generateAESKey()
			if err != nil {
				log.Println("Error generating AES key:", err)
				return nil
			}

			err = insertFileInfo(db, path, info.Name(), info.Size(), sha256Hash, aesKey)
			if err != nil {
				log.Println("Error inserting file info:", err)
			}
		}
		return nil
	})

	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("File traversal, hashing, and recording complete.")

	// Fetch and print unuploaded files
	unuploadedFiles, err := getUnuploadedFiles(db)
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println("Unuploaded files:")
	for _, file := range unuploadedFiles {
		fmt.Printf("ID: %d, Path: %s, Name: %s, Size: %d, SHA256: %s, AESKey: %s, Uploaded: %d\n",
			file.ID, file.Path, file.Name, file.Size, file.SHA256, file.AESKey, file.Uploaded)
	}

	//go func() {
	//	_, err := processUploads(db)
	//	if err != nil {
	//
	//	}
	//}()
	_, err = processUploads(db)
	if err != nil {
		log.Println("Error processing uploads:", err)
	}
}

// hashFileSHA256 hashes the file at the given path using SHA256
func hashFileSHA256(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hasher := sha256.New()
	if _, err := io.Copy(hasher, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hasher.Sum(nil)), nil
}

// generateAESKey generates a 32-byte AES key and returns it as a string
func generateAESKey() ([]byte, error) {
	key := make([]byte, 32)
	_, err := rand.Read(key)
	if err != nil {
		return nil, err
	}

	return key, nil
}

// insertFileInfo inserts the file information, hash, AES key, and default uploaded value into the database
func insertFileInfo(db *sql.DB, path, name string, size int64, sha256Hash string, aesKey []byte) error {
	insertQuery := `INSERT INTO files (path, name, size, sha256, aes_key, uploaded) VALUES (?, ?, ?, ?, ?, 0)`
	_, err := db.Exec(insertQuery, path, name, size, sha256Hash, aesKey)
	return err
}

// hashExistsInDB checks if the given SHA256 hash already exists in the database
func hashExistsInDB(db *sql.DB, sha256Hash string) (bool, error) {
	var exists bool
	query := `SELECT EXISTS(SELECT 1 FROM files WHERE sha256 = ? LIMIT 1)`
	err := db.QueryRow(query, sha256Hash).Scan(&exists)
	return exists, err
}

// getUnuploadedFiles queries the database for files where uploaded = 0
func getUnuploadedFiles(db *sql.DB) ([]FileInfo, error) {
	query := `SELECT id, path, name, size, sha256, aes_key, uploaded FROM files WHERE uploaded = 0`
	rows, err := db.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var files []FileInfo
	for rows.Next() {
		var file FileInfo
		err := rows.Scan(&file.ID, &file.Path, &file.Name, &file.Size, &file.SHA256, &file.AESKey, &file.Uploaded)
		if err != nil {
			return nil, err
		}
		files = append(files, file)
	}

	if err = rows.Err(); err != nil {
		return nil, err
	}

	return files, nil
}
