package main

import (
	"encoding/base64"
	"fmt"
	"log"
	"os"
	"os/exec"
	"regexp"
	"time"

	"github.com/go-co-op/gocron"
)

const (
	PostgreSQLURI      = "POSTGRESQL_URI"
	CronInterval       = "CRON_INTERVAL"
	ExternalBackup     = "EXTERNAL_BACKUP"
	ExternalBackupPath = "EXTERNAL_BACKUP_PATH"
	User               = "backupper"
	Group              = "backupp"
	Path               = "/backup/"
)

func main() {
	if os.Getenv(PostgreSQLURI) == "" {
		log.Fatalln(PostgreSQLURI + " is not defined")
		return
	}
	externalBackup := os.Getenv(ExternalBackup)
	if externalBackup != "" {
		decoded, err := base64.StdEncoding.DecodeString(externalBackup)
		if err == nil {
			mkdirCmd := exec.Command("runuser", "-u", User, "--", "bash", "-c", "mkdir -p /home/"+User+"/.config/rclone")
			mkdirCmd.Stdin = os.Stdin
			mkdirCmd.Stdout = os.Stdout
			mkdirCmd.Stderr = os.Stderr
			mkdirCmd.Run()
			os.WriteFile(fmt.Sprintf("/home/%s/.config/rclone/rclone.conf", User), decoded, 0644)
			permissionCmd := exec.Command("runuser", "-u", User, "--", "bash", "-c", fmt.Sprintf("chown %s:%s /home/%s/.config/rclone/rclone.conf && chmod 644 /home/%s/.config/rclone/rclone.conf", User, Group, User, User))
			permissionCmd.Stdin = os.Stdin
			permissionCmd.Stdout = os.Stdout
			permissionCmd.Stderr = os.Stderr
			permissionCmd.Run()
		}
	}

	dumpCmd := exec.Command("chown", User+":"+Group, Path)
	dumpCmd.Stdin = os.Stdin
	dumpCmd.Stdout = os.Stdout
	dumpCmd.Stderr = os.Stderr
	dumpCmd.Run()

	interval := os.Getenv(CronInterval)
	if interval != "" {
		re := regexp.MustCompile(`^(\*|([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])|\*\/([0-9]|1[0-9]|2[0-9]|3[0-9]|4[0-9]|5[0-9])) (\*|([0-9]|1[0-9]|2[0-3])|\*\/([0-9]|1[0-9]|2[0-3])) (\*|([1-9]|1[0-9]|2[0-9]|3[0-1])|\*\/([1-9]|1[0-9]|2[0-9]|3[0-1])) (\*|([1-9]|1[0-2])|\*\/([1-9]|1[0-2])) (\*|([0-6])|\*\/([0-6]))$`)
		if re.MatchString(interval) {
			fmt.Println("Starting cron with interval: " + interval)
			timezone := os.Getenv("TZ")
			if timezone == "" {
				timezone = "UTC"
			}
			loc, err := time.LoadLocation(timezone)
			if err != nil {
				loc = time.Local
			}
			s := gocron.NewScheduler(loc)
			s.Cron(interval).Do(runBackup)
			s.StartBlocking()
		} else {
			log.Fatal(CronInterval + " is not valid")
		}
	} else {
		runBackup()
	}
}

func runBackup() {
	timezone := os.Getenv("TZ")
	if timezone == "" {
		timezone = "UTC"
	}
	loc, _ := time.LoadLocation(timezone)
	time := time.Now().In(loc)

	filename := "backup-" + time.Format("01-02-2006-15-04-05") + ".sql"

	postgresqluri := os.Getenv(PostgreSQLURI)
	dumpCmd := exec.Command("runuser", "-u", User, "--", "bash", "-c", "pg_dump --dbname="+postgresqluri+" --file="+Path+filename)
	dumpCmd.Stdin = os.Stdin
	dumpCmd.Stdout = os.Stdout
	dumpCmd.Stderr = os.Stderr
	dumpCmd.Run()

	externalPath := os.Getenv(ExternalBackupPath)
	if externalPath != "" {
		backupCmd := exec.Command("runuser", "-u", User, "--", "bash", "-c", fmt.Sprintf("rclone copy %s%s remote:%s", Path, filename, externalPath))
		backupCmd.Stdin = os.Stdin
		backupCmd.Stdout = os.Stdout
		backupCmd.Stderr = os.Stderr
		backupCmd.Run()
	}
}
