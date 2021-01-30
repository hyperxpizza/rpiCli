package main

import (
	"bufio"
	"fmt"
	"net"
	"os/exec"
	"strings"

	"github.com/hyperxpizza/rpiCli/config"
	pb "github.com/hyperxpizza/rpiCli/grpc"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	pb.UnimplementedCommandServiceServer
}

func (s *Server) ExecuteCommand(request *pb.ExecuteCommandRequest, stream pb.CommandService_ExecuteCommandServer) error {
	// get bash command and split it into an array
	args := strings.Split(request.GetCommand(), " ")

	cmd := exec.Command(args[0], args...)

	stderr, err := cmd.StderrPipe()
	if err != nil {
		logrus.Printf("[-] StderrPipe error: %v\n", err)
		return err
	}
	cmd.Start()

	scanner := bufio.NewScanner(stderr)
	scanner.Split(bufio.ScanWords)

	for scanner.Scan() {
		message := scanner.Text()
		logrus.Println(message)

		response := pb.ExecuteCommandResponse{
			Response: message,
			Error:    nil,
		}

		if err := stream.Send(&response); err != nil {
			logrus.Printf("[-] Error while sending stream: %v\n", err)
		}
	}

	return nil
}

func main() {
	config := config.Init("")

	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", config.Server.Port))
	if err != nil {
		logrus.Fatalf("[-] Failed to listen: %v\n", err)
	}

	/*
		path := config.Server.CertPath
		creds, err := credentials.NewServerTLSFromFile(path+"/server-cert.pem", path+"/server-key.pem")
		if err != nil {
			log.Fatal("cannot load TLS credentials: ", err)
		}
	*/

	grpcServer := grpc.NewServer()
	pb.RegisterCommandServiceServer(grpcServer, &Server{})
	logrus.Printf("[+] Server running at: %s:%d", config.Server.Host, config.Server.Port)
	if err := grpcServer.Serve(lis); err != nil {
		logrus.Printf("[-] Failed to serve: %v", err)
	}
}
