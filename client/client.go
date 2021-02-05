package main

import (
	"bufio"
	"bytes"
	"context"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"path/filepath"

	"github.com/TwinProduction/go-color"
	"github.com/hyperxpizza/rpiCli/config"
	pb "github.com/hyperxpizza/rpiCli/grpc"
	"github.com/hyperxpizza/rpiCli/helpers"
	"github.com/prometheus/common/log"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var interactive *bool
var fileOutput *string
var fileInput *string
var fileUpload *string
var fileDownload *string
var savePath *string

func init() {
	interactive = flag.Bool("interactive", false, "If set, run client in an interactive mode.")
	fileOutput = flag.String("fileOutput", "", "Save output to file. Example: --fileOutput=example.txt")
	fileInput = flag.String("fileInput", "", "Run bash from input file. Example: --fileInput=script.sh")
	fileUpload = flag.String("fileUpload", "", "Upload file. Example: --fileUpload=/path/to/file")
	fileDownload = flag.String("fileDownload", "", "Download file. Example: --fileDownload=file.extension")
	savePath = flag.String("savePath", "", "Path for file saving. Example: --savePath=/path/to/destination/file.extension")

	flag.Parse()
}

func main() {
	config := config.Init("")

	/*
		path := config.Client.CertPath
		creds, err := credentials.NewServerTLSFromFile(path+"/server-cert.pem", path+"/server-key.pem")
		if err != nil {
			logrus.Fatalf("[-] Cannot load TLS credentials: %v\n", err)
		}*/

	address := fmt.Sprintf("%s:%d", config.Client.Host, config.Client.Port)
	/*
		connection, err := grpc.Dial(address, grpc.WithTransportCredentials(creds), grpc.WithBlock())
		if err != nil {
			logrus.Fatalf("[-] Error while creating connection")
		}
	*/
	connection, err := grpc.Dial(address, grpc.WithInsecure(), grpc.WithBlock())
	if err != nil {
		logrus.Fatalf("[-] Error while creating connection")
	}

	client := pb.NewCommandServiceClient(connection)
	if *interactive {
		interactiveCli(client, address)
	} else if *fileInput != "" {
		payload := helpers.LoadFile(*fileInput)
		sendPayload(client, payload)
	} else if *fileUpload != "" {
		upload(client, *fileUpload)
	} else if *fileDownload != "" {
		if *savePath == "" {
			log.Fatalln("Path for saving not specified")
		}

		downloadFile(client, *savePath, *fileDownload)
	}

}

func interactiveCli(client pb.CommandServiceClient, address string) {
	running := true
	scanner := bufio.NewScanner(os.Stdin)
	for running {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()

		var command string
		fmt.Print(color.Ize(color.Purple, address+":"))
		scanner.Scan()
		command = scanner.Text()
		request := pb.ExecuteCommandRequest{
			Command: command,
		}

		stream, err := client.ExecuteCommand(ctx, &request)
		if err != nil {
			logrus.Fatalf("[-] ExecuteCommand error: %v\n", err)
		}

		helpers.PrintStream(stream)

	}
}

func sendPayload(client pb.CommandServiceClient, payload string) {
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	fmt.Printf("Payload in sendPayload: %s\n", payload)

	fmt.Print(color.Ize(color.Green, fmt.Sprintf("[*] Sending payload: %s", payload)))
	request := pb.ExecuteCommandRequest{
		Command: payload,
	}

	stream, err := client.ExecuteCommand(ctx, &request)
	if err != nil {
		logrus.Fatalf("[-] ExecuteCommand error: %v\n", err)
	}

	helpers.PrintStream(stream)
}

func upload(client pb.CommandServiceClient, path string) {
	//open file
	file, err := os.Open(path)
	if err != nil {
		logrus.Fatalf("[-] Can not open file: %w\n", err)
	}
	defer file.Close()

	//get full size of file
	info, err := os.Stat("./court.png")
	if err != nil {
		logrus.Fatal(err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	stream, err := client.UploadFile(ctx)
	if err != nil {
		logrus.Fatalf("[-] Can not upload file: %w\n", err)
	}

	_, f1 := filepath.Split(path)
	fileName := strings.ReplaceAll(f1, filepath.Ext(path), "")

	request := &pb.UploadFileRequest{
		Data: &pb.UploadFileRequest_Info{
			Info: &pb.FileInfo{
				Filename:     fileName,
				Filetype:     filepath.Ext(path),
				FullFilesize: uint64(info.Size()),
			},
		},
	}

	err = stream.Send(request)
	if err != nil {
		logrus.Fatal("[-] Cannot send image info to server: ", err, stream.RecvMsg(nil))
	}

	reader := bufio.NewReader(file)
	buffer := make([]byte, 1024)

	for {
		n, err := reader.Read(buffer)
		if err == io.EOF {
			break
		}

		if err != nil {
			logrus.Fatalf("[-] Can not read chunk to buffer: %w", err)
		}

		request := &pb.UploadFileRequest{
			Data: &pb.UploadFileRequest_ChunkData{
				ChunkData: buffer[:n],
			},
		}

		err = stream.Send(request)

		if err != nil {
			logrus.Fatalf("[-] Can not send chunk to server: %w\n", err)
		}
	}

	response, err := stream.CloseAndRecv()
	if err != nil {
		logrus.Fatalf("[-] Can not recieve response: %w\n", err)
	}

	logrus.Printf(fmt.Sprintf("[+] %d bytes uploaded", response.Size))
}

func downloadFile(client pb.CommandServiceClient, destinationPath, fileName string) {
	logrus.Println("[*] Downloading...")

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	request := pb.DownloadFileRequest{
		Filename: fileName,
	}

	stream, err := client.DownloadFile(ctx, &request)
	if err != nil {
		logrus.Fatal(err)
	}

	fileData := bytes.Buffer{}

	for {
		response, err := stream.Recv()
		if err == io.EOF {
			break
		}

		if err != nil {
			logrus.Fatal(err)
		}

		logrus.Printf("[*] Recieved %d bytes\n", len(response.ChunkData))
		_, err = fileData.Write(response.ChunkData)
		if err != nil {
			logrus.Fatal(err)
		}
	}

	file, err := os.Create(destinationPath + "/" + fileName)
	if err != nil {
		logrus.Fatal(err)
	}

	_, err = fileData.WriteTo(file)
	if err != nil {
		logrus.Fatal(err)
	}

	logrus.Printf("[+] Downloading finished. File saved in: %s\n", destinationPath+"/"+fileName)
}
