# MariaDB Backup

This app automate creating backup from MariaDB database. Procedure includes these steps:  
* Create backup from database
* Encrypt backup
* Push it to object storage
* Remove old backups from object storage and local directory according to specified rotation policies
* Trigger alarm in slack channel about status of procedure

Application support two different types of backup:
* mysqldump
* mariabackup  
**Note**: This option can be specified by *BACKUP_TYPE* environment variable. All of the options can be seen in *docker-compose.yml* file.