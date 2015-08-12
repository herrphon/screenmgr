package server

import "log"

func Run() {
	checkError("Find devices", findDevices())
	checkError("Start HTTP server", runHTTPServer())
}

func checkError(action string, err error) {
	if err != nil {
		log.Fatal(action, " error: ", err.Error())
	}
}
