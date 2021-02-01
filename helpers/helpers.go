package helpers

import (
	"fmt"
	"io"
	"io/ioutil"
	"log"

	pb "github.com/hyperxpizza/rpiCli/grpc"
	"github.com/sirupsen/logrus"
)

func LoadFile(path string) string {
	content, err := ioutil.ReadFile(path)
	if err != nil {
		logrus.Fatal(err)
	}

	text := string(content)
	log.Println(text)
	return text
}

func PrintStream(stream pb.CommandService_ExecuteCommandClient) {
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
