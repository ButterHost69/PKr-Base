package dialer

import (
	// "ButterHost69/PKr-base/encrypt"
	// "ButterHost69/PKr-base/filetracker"
	// "ButterHost69/PKr-base/logger"
	// "ButterHost69/PKr-base/models"
	// "ButterHost69/PKr-base/pb"

	"github.com/ButterHost69/PKr-Base/config"
	"github.com/ButterHost69/PKr-Base/encrypt"
	"github.com/ButterHost69/PKr-Base/filetracker"
	"github.com/ButterHost69/PKr-Base/logger"

	// "github.com/ButterHost69/PKr-Base/models"
	"github.com/ButterHost69/PKr-Base/pb"

	"context"
	"fmt"
	"io/ioutil"
	"net"
	"os"
	"path"
	"strings"
	"time"

	"google.golang.org/grpc"
)

func getMyIP() (string, error) {
	conn, err := net.Dial("udp", "8.8.8.8:80")
	if err != nil {
		return "nil", err
	}
	defer conn.Close()

	localAddr := conn.LocalAddr().(*net.UDPAddr)
	return localAddr.IP.String(), nil
}

func getData(workspace_ip, workspace_path, workspace_name string) error {
	// fmt.Println("Connecting to Data Server ...") // [ ] Debug
	my_ip, err := getMyIP()
	if err != nil {
		return err
	}

	// [ ] : Send Port, from UserConfig File
	my_ip = my_ip + ":9000"

	conn, err := grpc.NewClient(workspace_ip, grpc.WithInsecure())
	if err != nil {
		return err
	}
	defer conn.Close()

	client := pb.NewDataServiceClient(conn)

	// [ ] Set Appropriate Time Limit
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	stream, err := client.GetData(ctx, &pb.DataRequest{
		WorkspaceName: workspace_name,
		ConnectionIp:  my_ip,
	})

	if err != nil {
		return err
	}

	var data_bytes []byte
	var key_bytes []byte
	var iv_bytes []byte

	var last_hash string
	for {
		data, err := stream.Recv()
		if err != nil {
			return err
		}

		if data.Filetype == 0 {
			data_bytes = append(data_bytes, data.Chunk...)
		} else if data.Filetype == 1 {
			key_bytes = append(key_bytes, data.Chunk...)
		} else if data.Filetype == 2 {
			iv_bytes = append(iv_bytes, data.Chunk...)
			break
		} else if data.Filetype == 3 {
			last_hash = string(data.Chunk)
		}
	}

	decrypted_key, err := encrypt.DecryptData(string(key_bytes))
	if err != nil {
		return err
	}

	decrypted_iv, err := encrypt.DecryptData(string(iv_bytes))
	if err != nil {
		return err
	}

	data, err := encrypt.AESDecrypt(data_bytes, decrypted_key, decrypted_iv)
	if err != nil {
		return err
	}

	zipFileName := strings.Split(time.Now().String(), " ")[0] + ".zip"
	zip_file_path := workspace_path + "\\.PKr\\" + zipFileName
	zippedfile, err := os.Create(zip_file_path)
	if err != nil {
		return err
	}
	defer zippedfile.Close()

	zippedfile.Write(data)

	// Delete all files in the GetWorkspace Dir except for .PKr
	files, err := ioutil.ReadDir(workspace_path)
	if err != nil {
		return err
	}

	// fmt.Printf("Deleting All Files at: %s\n\n", workspace_path)
	for _, file := range files {
		if file.Name() != ".PKr" && file.Name() != "PKr-base.exe" && file.Name() != "PKr-cli.exe" && file.Name() != "tmp" {
			if err = os.RemoveAll(path.Join([]string{workspace_path, file.Name()}...)); err != nil {
				return err
			}
		}
	}

	// Unzip Content
	// unzip_file_path := getworkspace.WorkspacePath
	if err = filetracker.UnzipData(zip_file_path, workspace_path+"\\"); err != nil {
		return err
	}

	if err = config.UpdateGetWorkspaceFolderToUserConfig(workspace_name, workspace_path, workspace_ip, last_hash); err != nil {
		return fmt.Errorf("error in adding GetConnection to the Main User Config Folder.\nerror:%v", err)
	}
	// fmt.Println("Data Transfer Completed ...") // [ ] Debug
	return nil
}

func PullData(userConfig_log *logger.UserLogger, workspace_name string) error {
	fmt.Println("Sending Pull Data Request ...") // [ ]: Debug
	user_config, err := config.ReadFromUserConfigFile()
	if err != nil {
		log_entry := fmt.Sprintf("cannot read from user config\nError: %s\nSource: PullData() Dialer", err.Error())
		// models.AddUsersLogEntry(log_entry)
		// fmt.Println(log_entry)
		userConfig_log.Debug(log_entry)
		return err
	}

	var workspace_ip, workspace_path string
	for _, get_workspace := range user_config.GetWorkspaces {
		// [ ] Get a Proper Unique Identifier instead of workspace_name
		if get_workspace.WorkspaceName == workspace_name {
			workspace_ip = get_workspace.WorkspcaceIP
			workspace_path = get_workspace.WorkspacePath
			break
		}
	}
	if workspace_ip == "" {
		log_entry := "invalid workspace name\nSource: PullData() Dialer"
		// models.AddUsersLogEntry(log_entry)
		// fmt.Println(log_entry)
		userConfig_log.Debug(log_entry)
		return fmt.Errorf("invalid workspace name")
	}

	conn, err := grpc.NewClient(workspace_ip, grpc.WithInsecure())
	if err != nil {
		log_entry := fmt.Sprintf("cannot connect to pull data handler\nError: %s\nSource: PullData() Dialer", err.Error())
		// models.AddUsersLogEntry(log_entry)
		// fmt.Println(log_entry)
		userConfig_log.Debug(log_entry)
		return err
	}
	defer conn.Close()

	client := pb.NewBackgroundServiceClient(conn)

	// [ ] Set Appropriate Time Limit
	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	res, err := client.PullData(ctx, &pb.PullDataRequest{
		WorkspaceName: workspace_name,
	})
	log :=  fmt.Sprintf("Assigned PORT:", res.Port)
	userConfig_log.Info(log)

	if err != nil {
		log_entry := fmt.Sprintf("error from server(pull data handler)\nError: %s\nSource: PullData() Dialer", err.Error())
		// models.AddUsersLogEntry(log_entry)
		// fmt.Println(log_entry)
		userConfig_log.Debug(log_entry)
		return err
	}

	port_int := int(res.Port)
	only_ip := strings.Split(workspace_ip, ":")[0]
	data_service_ip := fmt.Sprintf("%s:%d", only_ip, port_int)

	log = fmt.Sprintf("PORT:", port_int)
	userConfig_log.Info(log)

	log = fmt.Sprintf("Only IP: ", only_ip)
	userConfig_log.Info(log)

	log = fmt.Sprintf("IP Address:", data_service_ip)
	userConfig_log.Info(log)

	return getData(data_service_ip, workspace_path, workspace_name)
}
