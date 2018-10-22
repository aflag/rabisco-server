package rabisco

import (
	"fmt"
	"io/ioutil"
	"log"
	"math/rand"
	"os"
	"os/exec"
	"strconv"

	"github.com/mongodb/mongo-go-driver/mongo"
	"github.com/sirupsen/logrus/hooks/test"
)

func mongoServer() (*exec.Cmd, string, string) {
	dir, err := ioutil.TempDir("", "mongo")
	if err != nil {
		log.Fatal(err)
	}
	randomPort := strconv.Itoa(rand.Intn(10000) + 1000)
	url := fmt.Sprintf("mongodb://localhost:%s", randomPort)
	// nounixsocket prevents the test leaving socks behind in /tmp
	cmd := exec.Command("mongod", "--port", randomPort, "--dbpath", dir, "--quiet", "--nounixsocket")
	cmd.Start()
	return cmd, url, dir
}

func rabiscoServer() (r Backend, mongoURL string, cleanUp func()) {
	cmd, mongoURL, dir := mongoServer()
	cleanUp = func() {
		cmd.Process.Kill()
		os.RemoveAll(dir)
	}
	logger, _ := test.NewNullLogger()
	r, err := NewBackend(nil, logger, mongoURL)
	if err != nil {
		log.Fatal("Couldn't start rabisco")
	}
	return
}

func getRoomsColl(mongoURL string) *mongo.Collection {
	client, err := mongo.NewClient(mongoURL)
	if err != nil {
		log.Fatal("Error creating new client")
	}
	if client.Connect(nil) != nil {
		log.Fatal("Couldn't connect to mongo db")
	}
	return client.Database("rabisco").Collection("rooms")
}
