package s3

import (
	"bytes"
	"crypto/md5"
	"crypto/tls"
	"fmt"
	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
	"github.com/aws/aws-sdk-go/service/s3/s3manager"
	"io"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"

	"log"
	"os-database-backup/pkg/utl"
	"path/filepath"
	"time"
)

const (
	firstMonthDatePattern = "01T"
)

type RemoteBackupManager struct {
	s3Session *session.Session
	conf      *utl.S3Config
}

func NewRemoteBackupManager() (*RemoteBackupManager, error) {
	c := utl.GetS3Config()
	tr := &http.Transport{
		TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
	}
	client := &http.Client{Transport: tr}
	sess, err := session.NewSession(
		&aws.Config{
			Endpoint:   aws.String(c.Endpoint),
			Region:     aws.String(c.Region),
			HTTPClient: client,
			DisableSSL: aws.Bool(true),
			Credentials: credentials.NewStaticCredentials(
				c.AccessKeyID,
				c.SecretAccessKey,
				"",
			),
		})
	if err != nil {
		return nil, err
	}
	si := RemoteBackupManager{
		conf:      c,
		s3Session: sess,
	}
	return &si, nil
}

func (rbm *RemoteBackupManager) Execute(backupPath string) error {
	chunkSize, err := rbm.uploadS3Backup(backupPath, rbm.conf.Bucket)
	if err == nil {
		val, err := rbm.validateRemoteBackup(backupPath, chunkSize)
		if err == nil {
			if val {
				log.Println("Info s3  Backup is valid on s3")
				err = rbm.rotateRemoteBackup()
				if err != nil {
					log.Println("Error s3  Rotating remote backup is not successful")
				}
			} else {
				log.Println("Error s3  Backup is not valid on s3")
			}
		} else {
			log.Println("Error s3  Validating remote backup is not successful")
		}
	} else {
		log.Println("Error s3  Upload backup to s3 is not successful")
	}

	return err
}

func (rbm *RemoteBackupManager) uploadS3Backup(filePath, Bucket string) (int, error) {
	gzfile, gzerr := os.Open(filePath)
	if gzerr != nil {
		log.Println("Error s3  " + string(gzerr.Error()))
		return -1, gzerr
	}
	defer gzfile.Close()

	uploader := s3manager.NewUploader(rbm.s3Session)

	var startUpload = time.Now()
	log.Println("Info s3  Uploading backup to s3 starting...")
	_, err := uploader.Upload(&s3manager.UploadInput{
		Bucket: aws.String(Bucket),
		ACL:    aws.String("private"),
		Key:    aws.String(filepath.Base(filePath)),
		Body:   gzfile,
	})
	var endUpload = time.Now()
	if err != nil {
		log.Println("Error s3  " + string(err.Error()))
		return -1, err
	}

	var totalUploadTime = endUpload.Sub(startUpload)
	totalUploadTimeString := fmt.Sprintf("%v", totalUploadTime.Seconds())
	log.Println("Info s3  Backup uploaded to s3 successfully " + filepath.Base(filePath))
	log.Println("Info s3  TotalUploadedTime: " + totalUploadTimeString)

	return int(uploader.PartSize), nil
}

func (rbm *RemoteBackupManager) validateRemoteBackup(filePath string, chunkSize int) (bool, error) {
	svc := s3.New(rbm.s3Session, &aws.Config{
		Region:   aws.String(rbm.conf.Region),
		Endpoint: aws.String(rbm.conf.Endpoint),
	})
	input := &s3.HeadObjectInput{
		Bucket: aws.String(rbm.conf.Bucket),
		Key:    aws.String(filepath.Base(filePath)),
	}
	result, err := svc.HeadObject(input)
	if err != nil {
		log.Println("Error s3  " + err.Error())
		return false, err
	}

	tags := result.GoString()

	t, err := calculateMultipartEtag(filePath, chunkSize)
	if err != nil {
		return false, err
	}
	log.Println("Debug s3  ETag of protected backup is: " + t)

	if strings.Contains(tags, t) {
		log.Println("Info s3  Backup on object storage and database are the same")
		return true, nil
	} else {
		log.Println("Error s3  Backup on object storage and database are not the same!")
		return false, nil
	}

}

func (rbm *RemoteBackupManager) rotateRemoteBackup() error {
	svc := s3.New(rbm.s3Session, &aws.Config{
		Region:   aws.String(rbm.conf.Region),
		Endpoint: aws.String(rbm.conf.Endpoint),
	})

	resp, err := svc.ListObjectsV2(&s3.ListObjectsV2Input{Bucket: aws.String(rbm.conf.Bucket)})
	if err != nil {
		log.Println("Error s3  Can not list objects for deletion. " + err.Error())
		return err
	}

	var dcBackups []string
	var dcFMBackups []string
	for _, item := range resp.Contents {
		if strings.HasPrefix(*item.Key, rbm.conf.DC) && !strings.Contains(*item.Key, firstMonthDatePattern) {
			dcBackups = append(dcBackups, *item.Key)
		} else if strings.HasPrefix(*item.Key, rbm.conf.DC) && strings.Contains(*item.Key, firstMonthDatePattern) {
			dcFMBackups = append(dcFMBackups, *item.Key)
		}
	}

	var delete []string
	sort.Strings(dcBackups)
	sort.Strings(dcFMBackups)

	retentionNumberS3, _ := strconv.Atoi(rbm.conf.S3RetentionNumber)
	if len(dcBackups)-retentionNumberS3 >= 1 {
		for _, b := range dcBackups[:len(dcBackups)-retentionNumberS3] {
			delete = append(delete, b)
		}
	}

	FMRetentionNumberS3, _ := strconv.Atoi(rbm.conf.FMRetentionNumber)
	if len(dcFMBackups)-FMRetentionNumberS3 >= 1 {
		for _, b := range dcFMBackups[:len(dcFMBackups)-FMRetentionNumberS3] {
			delete = append(delete, b)
		}
	}

	for _, d := range delete {
		_, err = svc.DeleteObject(&s3.DeleteObjectInput{Bucket: aws.String(rbm.conf.Bucket), Key: aws.String(d)})
		if err != nil {
			log.Println("Error s3  Can not delete object " + d)
			return err
		}

		err = svc.WaitUntilObjectNotExists(&s3.HeadObjectInput{
			Bucket: aws.String(rbm.conf.Bucket),
			Key:    aws.String(d),
		})
		if err != nil {
			log.Println("Error s3  Problem while waiting for object " + d + " to be deleted")
			return err
		}

		log.Println("Info s3  Object " + d + " is successfully deleted on s3 because of remote rotation")
	}

	return nil
}

func calculateMultipartEtag(filePath string, chunk_size int) (string, error) {
	var md5s [][]byte
	f, err := os.Open(filePath)
	if err != nil {
		log.Println("Error s3  While opening file for etag calculation " + err.Error())
		return "", err
	}
	data := make([]byte, chunk_size)
	for {
		b, err := f.Read(data)
		if err == io.EOF {
			break
		}
		temp := md5.Sum(data[:b])
		md5s = append(md5s, temp[:])
	}
	if len(md5s) > 1 {
		return fmt.Sprintf("%x-%d", md5.Sum(bytes.Join(md5s, []byte(""))), len(md5s)), nil
	}
	if len(md5s) == 1 {
		return fmt.Sprintf("%x", md5s[0]), nil
	}
	return "", nil
}
