package dialer

import (
	"ButterHost69/PKr-base/logger"
	"ButterHost69/PKr-base/models"
	"ButterHost69/PKr-base/pb"
	"context"
	"fmt"
	"time"

	"google.golang.org/grpc"
)

// [ ]: Problem with IP Address
// Random PORT Address of Data Server is stored instead of PORT Address of Pkr-Base(Service)
func ScanForUpdatesOnStart(userConfig_log *logger.UserLogger) error {
	// Read all Get Workspaces from User Config & send request to check
	// whether there's new update or not

	user_config, err := models.ReadFromUserConfigFile()
	if err != nil {
		log_entry := "cannot read from workspace config file\nError: " + err.Error() + "\nSource: ScanForUpdatesOnStart() Dialer\n"
		fmt.Println(log_entry) // [ ]: Debugging
		// [ ]: Log Entry in Log file
		return err
	}

	for _, getWorkspace := range user_config.GetWorkspaces {
		conn, err := grpc.NewClient(getWorkspace.WorkspcaceIP, grpc.WithInsecure())
		if err != nil {
			log_entry := fmt.Sprintf("error while scanning for updates on start with %#v \nError: %s\nSource: ScanForUpdatesOnStart() Dialer\n", getWorkspace, err.Error())
			fmt.Println(log_entry) // [ ]: Debugging
			// [ ]: Log Entry in Log file
			continue
		}

		client := pb.NewBackgroundServiceClient(conn)

		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel() // [ ] Check if we can call cancel func instead of defering it, similar to conn.close()

		res, err := client.ScanForUpdatesOnStart(ctx, &pb.ScanForUpdatesRequest{
			WorkspaceName: getWorkspace.WorkspaceName,
			LastHash:      getWorkspace.LastHash,
		})

		if err != nil {
			if err == context.DeadlineExceeded {
				log_entry := fmt.Sprintf("error from sender while scanning for updates on start with %#v \nError: %s\nSource: ScanForUpdatesOnStart() Dialer", getWorkspace, "Context Deadline Exceeded, NO RESPONSE FROM SERVER, INVALID ADDRESS(MAYBE)\n")
				fmt.Println(log_entry) // [ ]: Debugging
				// [ ]: Log Entry in Log file
				continue
			}
			log_entry := fmt.Sprintf("error from sender while scanning for updates on start with %#v \nError: %s\nSource: ScanForUpdatesOnStart() Dialer\n", getWorkspace, err.Error())
			fmt.Println(log_entry) // [ ]: Debugging
			// [ ]: Log Entry in Log file
			continue
		}

		if res.NewUpdates {
			fmt.Printf("New Data is Available: %#v\n", getWorkspace) // [ ]: Debugging
			err := PullData(userConfig_log, getWorkspace.WorkspaceName)

			if err != nil {
				log_entry := fmt.Sprintf("error while pulling new updates from %#v \nError: %s\nSource: ScanForUpdatesOnStart() Dialer\n", getWorkspace, err.Error())
				fmt.Println(log_entry) // [ ]: Debugging
				// [ ]: Log Entry in Log file
				continue
			}
			continue
		}
		fmt.Printf("No New Data is Available: %#v\n", getWorkspace) // [ ]: Debugging

		// NOT USING defer cuz defer's gonna end connection at the end of the function
		// Instead we can end connection now
		conn.Close()
	}
	return nil
}
