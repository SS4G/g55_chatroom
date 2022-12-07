package main

import (
	"g55.com/chat/client"
	log "github.com/sirupsen/logrus"
	"os"
)

// 会在log 库中某处默认调用
func init() {
	// Log as JSON instead of the default ASCII formatter.
	//log.SetFormatter(&log.JSONFormatter{})

	// Output to stdout instead of the default stderr
	// Can be any io.Writer, see below for File example
	log.SetOutput(os.Stdout)

	// Only log the warning severity or above.
	log.SetLevel(log.DebugLevel)
}

func main() {
	//util.ClientTestMain()
	roomClient := client.NewChatRoomClient()
	roomClient.ClientRun()
	//model.ClientTestMain() //ModelTestRun()
}
