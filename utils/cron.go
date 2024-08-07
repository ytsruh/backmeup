package utils

import (
	"fmt"
	"os"

	"github.com/robfig/cron/v3"
)

func StartCronJobs() {
	c := cron.New()

	// Run every day at 00:00 - clean up temp zip directory
	c.AddFunc("0 0 * * *", func() {
		fmt.Println("Cleaning up temp zip directory")
		err := os.RemoveAll("zips")
		if err != nil {
			fmt.Printf("error removing zips directory: %s", err)
			return
		}
		// Crete a new zips directory
		err = os.MkdirAll("zips", 0755) // 0755 is the file permission (read and write permission)
		if err != nil {
			fmt.Printf("error creating zips directory: %s", err)
			return
		}
		fmt.Println("Successfully cleaned up temp zip directory")
	})

	c.Start()
	cronCount := len(c.Entries())
	fmt.Printf("setup %d cron jobs \n", cronCount)
}
