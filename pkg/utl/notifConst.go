package utl

type BackupStatus int
type NotifMessage string

const (
	FAILED       BackupStatus = 0
	SUCCEEDED    BackupStatus = 1
	FAILEDMSG    NotifMessage = "Backup procedure is not successful"
	SUCCEEDEDMSG NotifMessage = "Backup procedure is successful"
)
