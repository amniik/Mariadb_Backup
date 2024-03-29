package main

import (
	"log"
	"os"
	"os-database-backup/pkg/backup"
	"os-database-backup/pkg/s3"
	"os-database-backup/pkg/utl"
	"path/filepath"
	"time"
)

func main() {
	sn := utl.NewSlackNotif()
	bm, err := backup.NewBackupManager()
	if err != nil {
		log.Println("Error main  Creating backup manager is not successful")
		sn.SendSlackNotification(utl.FAILED, "", err.Error())
		os.Exit(4)
	}
	zipFile, err := bm.Execute()
	if err == nil {
		rbm, err := s3.NewRemoteBackupManager()
		if err != nil {
			log.Println("Error main  Creating remote backup manger is not successful")
			sn.SendSlackNotification(utl.FAILED, "", err.Error())
			os.Exit(4)
		}
		err = rbm.Execute(zipFile)
		if err != nil {
			log.Println("Error main  Executing remote backup is not successful")
			sn.SendSlackNotification(utl.FAILED, "", err.Error())
			os.Exit(4)
		}
	} else {
		log.Println("Error main  Executing backup is not successful")
		sn.SendSlackNotification(utl.FAILED, "", err.Error())
		os.Exit(4)
	}
	sn.SendSlackNotification(utl.SUCCEEDED, filepath.Base(zipFile), "")
}
