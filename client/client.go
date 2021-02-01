package main

import (
	"bufio"
	"context"
	"flag"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"time"

	"github.com/TwinProduction/go-color"
	"github.com/hyperxpizza/rpiCli/config"
	pb "github.com/hyperxpizza/rpiCli/grpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

var interactive *bool
var fileOutput *string
var fileInput *string

func init() {
	interactive = flag.Bool("interactive", false, "If set, run client in an interactive mode.")
	fileOutput = flag.String("fileOutput", "", "Save output to file. Example: --fileOutput=example.txt")
	fileInput = flag.String("fileInput", "", "Run bash from input file. Example: --fileInput=script.sh")

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

		for {
			response, err := stream.Recv()
			if err == io.EOF {
				break
			}

			if err != nil {
				logrus.Fatal(err)
			}

			fmt.Println(response.Response)
		}

	}
}

func loadFile(path string) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatal(err)
	}

	text := string(content)
	return text
}
