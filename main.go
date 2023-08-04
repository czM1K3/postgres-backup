package main

import (
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
	if _, err := os.Stat("/home/" + User + "/.config/rclone/rclone.conf"); err == nil {
		permissionCmd1 := exec.Command("chown", User+":"+Group, "/home/"+User+"/.config/rclone/rclone.conf")
		permissionCmd1.Stdin = os.Stdin
		permissionCmd1.Stdout = os.Stdout
		permissionCmd1.Stderr = os.Stderr
		permissionCmd1.Run()
		permissionCmd2 := exec.Command("chmod", "644", "/home/"+User+"/.config/rclone/rclone.conf")
		permissionCmd2.Stdin = os.Stdin
		permissionCmd2.Stdout = os.Stdout
		permissionCmd2.Stderr = os.Stderr
		permissionCmd2.Run()
	}

	permissionCmd := exec.Command("chown", User+":"+Group, "/home/"+User+"/.config/rclone")
	permissionCmd.Stdin = os.Stdin
	permissionCmd.Stdout = os.Stdout
	permissionCmd.Stderr = os.Stderr
	permissionCmd.Run()

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
			loc := getLocation()
			s := gocron.NewScheduler(&loc)
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
	loc := getLocation()
	time := time.Now().In(&loc)
	fmt.Printf("Starting backup at %s\n", time.String())

	filename := "backup-" + time.Format("01-02-2006-15-04-05") + ".sql"

	postgresqluri := os.Getenv(PostgreSQLURI)
	dumpCmd := exec.Command("runuser", "-u", User, "--", "bash", "-c", "pg_dump --dbname="+postgresqluri+" --file="+Path+filename)
	dumpCmd.Stdin = os.Stdin
	dumpCmd.Stdout = os.Stdout
	dumpCmd.Stderr = os.Stderr
	dumpCmd.Run()

	externalPath := os.Getenv(ExternalBackupPath)
	if externalPath != "" {
		fmt.Println("Sending backup with rclone")
		backupCmd := exec.Command("runuser", "-u", User, "--", "bash", "-c", fmt.Sprintf("rclone copy %s%s remote:%s", Path, filename, externalPath))
		backupCmd.Stdin = os.Stdin
		backupCmd.Stdout = os.Stdout
		backupCmd.Stderr = os.Stderr
		backupCmd.Run()
		fmt.Println("Finished sending backup")
	}
	fmt.Println("Exiting backup")
}

func getLocation() time.Location {
	timezone := os.Getenv("TZ")
	if timezone == "" {
		timezone = "UTC"
	}
	loc, err := time.LoadLocation(timezone)
	if err == nil {
		return *time.Local
	} else {
		return *loc
	}
}
