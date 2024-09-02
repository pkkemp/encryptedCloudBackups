package main

import (
	"database/sql"
	"log"
)

func processUploads(db *sql.DB) (bool, error) {
	files, _ := getUnuploadedFiles(db)
	for _, file := range files {
		success, err := uploadEncryptedFile("starlight-fusion", file.Path, []byte(file.AESKey), file.Path)
		if success {
			setUploadedError := setUploaded(db, file.ID)
			if setUploadedError != nil {
				log.Print("Error setting upload status in database")
				return false, setUploadedError
			}
		}
		if err != nil {
			log.Println(err.Error())
			return false, err
		}
	}
	return true, nil
}
