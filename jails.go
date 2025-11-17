/*
jails - An open-source small tool written in GO for automating backups of FreeBSD jails managed with Bastille.

To compile for FreeBSD execute "env GOOS=freebsd GOARCH=amd64 go build -o jails"
*/

package main

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"time"
)

var Version string = "0.6.20220911"
var Jails = []string{}
var InfoLogger *log.Logger
var ErrorLogger *log.Logger
var Days = "2"
var Mode = "live"
var Args int = len(os.Args)

func main() {

	if Args == 1 {

		toBackup(Mode, Days)

	} else if Args == 3 {

		flag1 := os.Args[1]
		flag2 := os.Args[2]

		if _, err := strconv.Atoi(flag2); err == nil {

			switch flag1 {

			case "safe":

				toBackup(flag1, flag2)

			case "live":

				toBackup(flag1, flag2)

			default:

				fmt.Println("")
				fmt.Println("Incorrect backup mode, please enter a correct value.")
				usageInfo()
			}

		} else {

			fmt.Println("")
			fmt.Println("Incorrect value for backup retention, please enter a number.")
			usageInfo()
		}

	} else {

		fmt.Println("")
		fmt.Println("Command incorrect or incomplete")
		usageInfo()
	}
}

func toBackup(Mode string, Days string) {

	// Collect jail names
	//arg := "bastille list | awk '{print $1}' | tail +2"   ### To collect running containers only ###
	arg := "bastille list containers"

	collect := exec.Command("csh", "-c", arg)

	stdout, err := collect.StdoutPipe()

	if err != nil {

		log.Fatal(err)
		ErrorLogger.Printf("Error: %s", err)
	}

	collect.Start()

	result := bufio.NewScanner(stdout)

	for result.Scan() {

		line := result.Text()
		Jails = append(Jails, line)
	}

	if len(Jails) == 0 {

		InfoLogger.Printf("----------------------------------------------------------------\n")
		InfoLogger.Printf("Starting Quimby - version %v (%v days of data retention)\n", Version, Days)
		log.Printf("Starting Quimby v%v (%v days of data retention)\n", Version, Days)
		log.Printf("No jails or Bastille found on this system! Check your jails and if Bastille is installed!")
		InfoLogger.Printf("No jails or Bastille found on this system! Check your jails and if Bastille is installed!")
		os.Exit(0)

	}

	// Backup jails
	InfoLogger.Printf("----------------------------------------------------------------\n")
	InfoLogger.Printf("Starting jails - version %v (%v days of data retention)\n", Version, Days)
	log.Printf("Starting jails v%v (%v days of data retention)\n", Version, Days)

	for _, i := range Jails {

		if Mode == "safe" {

			safeBackup(i)

		} else if Mode == "live" {

			liveBackup(i)
		}
	}

	// Remove backups older than X days
	arg1 := fmt.Sprintf("find /usr/local/bastille/backups/ -mtime +%v -print", Days)

	findFiles := exec.Command("csh", "-c", arg1)

	findFilesOut, err := findFiles.CombinedOutput()

	if err != nil {
		log.Fatal(err)
		ErrorLogger.Printf("Error: %s", err)
	}

	if len(findFilesOut) > 0 {

		InfoLogger.Printf("Removing backups older than %v days..\n", Days)
		log.Printf("Removing backups older than %v days..\n", Days)

		time.Sleep(10 * time.Second)

		arg2 := fmt.Sprintf("find /usr/local/bastille/backups/ -mtime +%v -delete;", Days)

		rmFiles := exec.Command("csh", "-c", arg2)

		_, err := rmFiles.CombinedOutput()

		if err != nil {
			log.Fatal(err)
			ErrorLogger.Printf("Error: %s", err)
		}

	} else {

		InfoLogger.Printf("No backups older than %v days found in folder...skipping cleanup.\n", Days)
		log.Printf("No backups older than %v days found in folder...skipping cleanup.\n", Days)
	}

	InfoLogger.Println("Backup Completed!")
	InfoLogger.Printf("----------------------------------------------------------------\n")
	log.Println("Backup Completed!")
}

// Logging configuration
func init() {
	file, err := os.OpenFile("/var/log/quimby.log", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0666)
	if err != nil {
		log.Fatal(err)
	}

	InfoLogger = log.New(file, "INFO: ", log.Ldate|log.Ltime|log.Lmsgprefix)
	ErrorLogger = log.New(file, "ERROR: ", log.Ldate|log.Ltime|log.Lshortfile|log.Lmsgprefix)
}

// Live Backup
func liveBackup(i string) {

	checkFS()

	backup := exec.Command("bastille", "export", "--gz", i)

	_, err := backup.StdoutPipe()

	if err != nil {

		fmt.Println(err)
		ErrorLogger.Printf("Error: %s", err)

	}

	InfoLogger.Printf("Hot backing up %v..\n", i)
	log.Printf("Hot backing up %v..\n", i)
	backup.Run()
	InfoLogger.Printf("Succesfully backed up %v!\n", i)
	log.Printf("Successfully backed up %v!\n", i)
}

// Safe Backup
func safeBackup(i string) {

	backup := exec.Command("bastille", "export", "--tgz", i)

	_, err := backup.StdoutPipe()

	if err != nil {

		fmt.Println(err)
		ErrorLogger.Printf("Error: %s", err)

	}

	InfoLogger.Printf("Stopping jail %v..\n", i)
	log.Printf("Stopping jail %v..\n", i)

	JailStop := exec.Command("bastille", "stop", i)
	JailStop.Run()

	InfoLogger.Printf("Backing up %v..\n", i)
	log.Printf("Backing up %v..\n", i)

	backup.Run()

	InfoLogger.Printf("Starting jail %v..\n", i)
	log.Printf("Starting jail %v..\n", i)

	JailStart := exec.Command("bastille", "start", i)
	JailStart.Run()

	InfoLogger.Printf("Succesfully backed up %v!\n", i)
	log.Printf("Successfully backed up %v!\n", i)
}

// Usage message
func usageInfo() {

	fmt.Println("")
	fmt.Println("Usage: quimby < safe | live > < days >")
	fmt.Println("")
	fmt.Println("Note: If no options are specified, the backup will run in live mode with 2 days of retention.")
	fmt.Println("")
}

// Check if running on ZFS filesystem
func checkFS() {

	checkfs := exec.Command("zpool", "list")

	checkfsout, err := checkfs.StdoutPipe()

	if err != nil {

		fmt.Println(err)
		ErrorLogger.Printf("Error: %s", err)
	}

	checkfs.Start()

	fstype := bufio.NewScanner(checkfsout)

	for fstype.Scan() {

		fsline := fstype.Text()

		if fsline == "no pools available" {

			log.Printf("This system is not running on ZFS, only safe backups will work.")
			InfoLogger.Printf("This system is not running on ZFS, only safe backups will work.")
			os.Exit(0)
		}
	}
}
