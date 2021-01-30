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

	arg1, arg2 := strings.Join(args[:1], " "), strings.Join(args[1:], " ")

	logrus.Println(args)

	cmd := exec.Command(arg1, arg2)
	r, _ := cmd.StdoutPipe()
	cmd.Stderr = cmd.Stdout

	done := make(chan struct{})

	scanner := bufio.NewScanner(r)

	go func() {
		for scanner.Scan() {
			line := scanner.Text()
			logrus.Println(line)

			message := pb.ExecuteCommandResponse{
				Response: line,
				Error:    nil,
			}

			if err := stream.Send(&message); err != nil {
				logrus.Println("[-] Error while sending the stream: %v", err)
			}
		}

		done <- struct{}{}
	}()

	err := cmd.Start()
	if err != nil {
		logrus.Println("cmd start error")
		logrus.Println(err)
		return err
	}

	<-done

	err = cmd.Wait()
	if err != nil {
		logrus.Println("cmd wait error")
		logrus.Println(err)
		return err
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
			log.Fatalf("[-] Cannot load TLS credentials: %v\n", err)
		}

		grpcServer := grpc.NewServer(
			grpc.Creds(creds),
		)
	*/
	grpcServer := grpc.NewServer()
	pb.RegisterCommandServiceServer(grpcServer, &Server{})
	logrus.Printf("[+] Server running at: %s:%d", config.Server.Host, config.Server.Port)
	if err := grpcServer.Serve(lis); err != nil {
		logrus.Printf("[-] Failed to serve: %v", err)
	}
}
