package main

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net"
	"os/exec"
	"strings"

	"github.com/hyperxpizza/rpiCli/config"
	pb "github.com/hyperxpizza/rpiCli/grpc"
	"github.com/hyperxpizza/rpiCli/server/filestorage"
	"github.com/sirupsen/logrus"
	"google.golang.org/grpc"
)

type Server struct {
	//pb.UnimplementedCommandServiceServer
	*filestorage.FileStorage
}

const maxFileSize = 1 << 20

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
				logrus.Println(err)
				if err.Error() == "rpc error: code = Unavailable desc = transport is closing" {
					return
				}
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

func (s *Server) UploadFile(stream pb.CommandService_UploadFileServer) error {
	request, err := stream.Recv()
	if err != nil {
		return err
	}

	fileName := request.GetInfo().GetFilename()
	fileType := request.GetInfo().GetFiletype()

	logrus.Printf(fmt.Sprintf("Recieving file: %s.%s", fileName, fileType))

	fileData := bytes.Buffer{}
	fileSize := 0

	logrus.Println("[*] Recieving data...")
	for {
		request, err := stream.Recv()
		if err == io.EOF {
			logrus.Println("[*] No more data")
			break
		}

		if err != nil {
			return err
		}

		chunk := request.GetChunkData()
		size := len(chunk)

		logrus.Println(fmt.Sprintf("[+] Recieved a chunk with size: %d", size))

		fileSize += size

		if fileSize > maxFileSize {
			return fmt.Errorf(fmt.Sprintf("File is too large: %d > %d", size, maxFileSize))
		}

		_, err = fileData.Write(chunk)
		if err != nil {
			return err
		}
	}

	name, err := s.FileStorage.Save(fileName, fileType, fileData)
	if err != nil {
		return err
	}

	response := &pb.UploadFileResponse{
		Id:    name,
		Size:  uint32(fileSize),
		Error: nil,
	}

	err = stream.SendAndClose(response)
	if err != nil {
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
	pb.RegisterCommandServiceServer(grpcServer, &Server{
		filestorage.NewFileStorage("/uploads"),
	})
	logrus.Printf("[+] Server running at: %s:%d", config.Server.Host, config.Server.Port)
	if err := grpcServer.Serve(lis); err != nil {
		logrus.Printf("[-] Failed to serve: %v", err)
	}
}
