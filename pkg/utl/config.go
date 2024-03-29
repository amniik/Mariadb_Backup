package utl

import (
	"os"
)

type BackupConfig struct {
	HostName             string
	Port                 string
	UserName             string
	Password             string
	DC                   string
	LocalDirectory       string
	BackupType           string
	ZipPass              string
	LocalRetentionNumber string
}

type S3Config struct {
	AccessKeyID       string
	SecretAccessKey   string
	Region            string
	Bucket            string
	Endpoint          string
	DC                string
	S3RetentionNumber string
	FMRetentionNumber string
}

type SlackConfig struct {
	SlackChannelName string
	SlackWebhookUrl  string
	BackupType       string
	DC               string
}

func GetBackupConfig() *BackupConfig {
	b := BackupConfig{
		HostName:             os.Getenv("DB_HOSTNAME"),
		Port:                 os.Getenv("DB_PORT"),
		UserName:             os.Getenv("DB_USERNAME"),
		Password:             os.Getenv("DB_PASSWORD"),
		DC:                   os.Getenv("DC"),
		LocalDirectory:       os.Getenv("LOCAL_DIRECTORY"),
		BackupType:           os.Getenv("BACKUP_TYPE"),
		ZipPass:              os.Getenv("ZIP_PASS"),
		LocalRetentionNumber: os.Getenv("LOCAL_RETENTION_NUMBER"),
	}
	return &b
}

func GetS3Config() *S3Config {
	s := S3Config{
		AccessKeyID:       os.Getenv("AWS_ACCESS_KEY_ID"),
		SecretAccessKey:   os.Getenv("AWS_SECRET_ACCESS_KEY"),
		Region:            os.Getenv("AWS_REGION"),
		Bucket:            os.Getenv("BUCKET_NAME"),
		Endpoint:          os.Getenv("AR_ENDPOINT"),
		DC:                os.Getenv("DC"),
		S3RetentionNumber: os.Getenv("S3_RETENTION_NUMBER"),
		FMRetentionNumber: os.Getenv("FIRST_MONTH_RETENTION_NUMBER"),
	}
	return &s
}

func GetSlackConfig() *SlackConfig {
	s := SlackConfig{
		SlackChannelName: "#" + os.Getenv("SLACK_CHANNEL_NAME"),
		SlackWebhookUrl:  os.Getenv("SLACK_WEBHOOK_URL"),
		BackupType:       os.Getenv("BACKUP_TYPE"),
		DC:               os.Getenv("DC"),
	}
	return &s
}
