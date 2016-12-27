package main

import (
	"flag"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/websocket"
	"github.com/ystv/nexus-common"

	"net/url"

	log "github.com/sirupsen/logrus"
)

func main() {
	args := struct {
		name, clientaddr, host string
		secureWebsockets       bool
	}{}
	flag.StringVar(&args.name, "name", "", "The stream's name")
	flag.StringVar(&args.clientaddr, "clientaddr", "", "IP address of stream's origin")
	flag.StringVar(&args.host, "server", "", "Nexus server host")
	flag.BoolVar(&args.secureWebsockets, "secure", true, "Use secure websockets")
	flag.Parse()

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT)
	t := time.NewTicker(time.Second * 1) // TODO: make configurable
	defer t.Stop()

	scheme := "wss"
	if !args.secureWebsockets {
		scheme = "ws"
	}
	u := url.URL{Scheme: scheme, Host: args.host, Path: "/v1/ws/streamstatus"}

	conn, _, err := websocket.DefaultDialer.Dial(u.String(), nil)
	if err != nil {
		log.Fatal(err)
	}
	defer conn.Close()

	go func() {
		for {
			if _, _, err := conn.NextReader(); err != nil {
				conn.Close()
				break
			}
		}
	}()

MainLoop:
	for {
		select {
		case <-t.C:
			err := conn.WriteJSON(&nexus_common.StreamUpdate{
				args.name,
				args.clientaddr,
				nexus_common.StreamStatusOnline,
			})
			if err != nil {
				log.Fatal(err)
			}

		case <-signalChan:
			// TODO: faster timeout for WriteJson to avoid too long a delay before exiting?
			err := conn.WriteJSON(&nexus_common.StreamUpdate{
				args.name,
				args.clientaddr,
				nexus_common.StreamStatusTerminating,
			})
			if err != nil {
				log.Fatal(err)
			}
			err = conn.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(websocket.CloseGoingAway, "Stream terminated, goodbye"))
			if err != nil {
				log.Fatal(err)
			}
			log.Info("Received signal, exiting")
			break MainLoop
		}
	}
}
