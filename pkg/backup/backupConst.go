package backup

type BackupOpt string
type BackupPostfix string
type ZipPostfix string

const (
	MysqlDump          BackupOpt     = "mysqldump"
	MariaBackup        BackupOpt     = "mariabackup"
	MysqlDumpPostfix   BackupPostfix = ".dump.sql"
	MariaBackupPostfix BackupPostfix = ".qp.xbc.xbs"
	BackupZipPostfix   ZipPostfix    = ".compress"
)
