package backup

import (
	"errors"
	"fmt"
	"github.com/alexmullins/zip"
	"io"
	"log"
	"os"
	"os-database-backup/pkg/utl"
	"os/exec"
	"path"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"
)

type BackupManager struct {
	option    string
	bpPostfix string
	conf      *utl.BackupConfig
}

func NewBackupManager() (*BackupManager, error) {
	c := utl.GetBackupConfig()
	var postfix string
	if c.BackupType == string(MysqlDump) {
		postfix = string(MysqlDumpPostfix)
	} else if c.BackupType == string(MariaBackup) {
		postfix = string(MariaBackupPostfix)
	} else {
		log.Println("Error backup  " + c.BackupType + " option is not supported")
		return nil, errors.New("option for backup manager is not supported")
	}
	bm := BackupManager{
		option:    c.BackupType,
		bpPostfix: postfix,
		conf:      c,
	}
	return &bm, nil
}

func (bm *BackupManager) Execute() (string, error) {
	startTime := time.Now()
	var backupFile string
	var zipFile string
	var err error
	if bm.option == string(MysqlDump) {
		backupPath := getBackupPath(startTime, bm.conf.DC, bm.conf.LocalDirectory, bm.bpPostfix)
		backupFile, err = dumpBackup(backupPath, startTime)
	} else if bm.option == string(MariaBackup) {
		backupPath := getBackupPath(startTime, bm.conf.DC, bm.conf.LocalDirectory, bm.bpPostfix)
		backupFile, err = mariaBackup(backupPath)
	} else {
		backupFile, err = "", errors.New("Function does not implemented for "+bm.option)
	}

	if err == nil {
		zipFile, err = zipProtectCompress(backupFile, bm.conf.ZipPass)
		if err != nil {
			log.Println("Error backup  Compressing backup is not successful")
		}
	}

	errl := rotateLocalBackups(bm.conf.LocalDirectory, bm.conf.LocalRetentionNumber)
	if errl != nil {
		log.Println("Error main  Rotating local backup is not successful")
	}

	if err != nil {
		return zipFile, err
	}
	return zipFile, errl
}

func getBackupPath(startTime time.Time, dc, dir, postfix_name string) string {
	prefix_name := dc + "_openstack-backup"
	start_time := strings.Replace(startTime.Format("2006-01-02T15:04:05-0700"), "-", "", -1)
	timestamp := strings.Replace(start_time, ":", "", -1)
	filename := prefix_name + "-" + timestamp + postfix_name
	filepath := path.Join(dir, filename)
	_ = os.Mkdir(path.Dir(filepath), os.ModePerm)

	return filepath
}

func getFileSize(filename string) (int64, error) {

	fi, err := os.Stat(filename)
	if err != nil {
		log.Println("Error backup  " + string(err.Error()))
		return 0, err
	}
	size := fi.Size()
	return size / 1024 / 1024, nil
}

func zipProtectCompress(filePath, zipPass string) (string, error) {
	content, err := os.Open(filePath)
	if err != nil {
		log.Println("Error backup  " + string(err.Error()))
		return "", err
	}

	save_path := filePath + string(BackupZipPostfix)
	fzip, err := os.Create(save_path)
	if err != nil {
		log.Println("Error backup  " + string(err.Error()))
		return "", err
	}

	zipw := zip.NewWriter(fzip)
	defer zipw.Close()
	w, err := zipw.Encrypt(save_path, zipPass)
	if err != nil {
		log.Println("Error backup  " + string(err.Error()))
		return "", err
	}
	_, err = io.Copy(w, content)
	if err != nil {
		log.Println("Error backup  " + string(err.Error()))
		return "", err
	}
	zipw.Flush()
	return save_path, nil
}

func dumpBackup(backupPath string, startTime time.Time) (string, error) {
	cmd := exec.Command("/mysqldump.sh", backupPath)
	log.Println("Info backup  Mysqldump is being executed...")
	_, err := cmd.Output()
	if err != nil {
		log.Println("Error backup  Mysqldump error: " + err.Error())
		return "", err
	}

	var backupTime = time.Now().Sub(startTime)
	backupFileSize, err := getFileSize(backupPath)
	if err != nil {
		log.Println("Error backup  " + err.Error())
		return "", err
	}
	backupFileSizeString := fmt.Sprintf("BackupSize: %vMB", float64(backupFileSize))
	backupTimeinSecond := fmt.Sprintf("BackupTime: %v ", backupTime)
	filename := filepath.Base(backupPath)
	log.Println("Info backup  Backup executed successfully " + filename)
	log.Println("Info backup  " + backupTimeinSecond + backupFileSizeString)

	return backupPath, nil
}

func mariaBackup(backupPath string) (string, error) {
	cmd := exec.Command("/mariabackup.sh", backupPath)
	log.Println("Info backup  Mariabackup is being executed...")
	_, err := cmd.Output()
	if err != nil {
		log.Println("Error backup  Mariabackup error: " + err.Error())
		return "", err
	}

	return backupPath, nil
}

func rotateLocalBackups(localDirectory, localRetentionNumber string) error {
	entries, _ := os.ReadDir(localDirectory)
	var remove []string
	var backups []string
	for _, e := range entries {
		if !strings.Contains(e.Name(), string(BackupZipPostfix)) {
			remove = append(remove, e.Name())
		} else {
			backups = append(backups, e.Name())
			log.Println("Debug backup  " + e.Name() + " exists on local")
		}
	}
	sort.Strings(backups)

	retention_number, err := strconv.Atoi(localRetentionNumber)
	if err != nil {
		log.Println("Error backup  Local retention number can not be converted to integer")
		return err
	}
	if len(backups)-retention_number >= 1 {
		for _, b := range backups[:len(backups)-retention_number] {
			remove = append(remove, b)
		}
	}
	for _, d := range remove {
		p := path.Join(localDirectory, d)
		err := os.Remove(p)
		if err != nil {
			log.Println("Error backup  " + string(err.Error()))
			return err
		}
		if strings.Contains(p, string(BackupZipPostfix)) {
			log.Println("Info backup  " + p + " is deleted because of backup rotation")
		} else {
			log.Println("Debug backup  Extra file " + p + " is deleted")
		}
	}

	return nil
}
