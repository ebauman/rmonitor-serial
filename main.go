package main

import (
	"fmt"
	"fyne.io/fyne/app"
	"fyne.io/fyne/widget"
	"github.com/tarm/serial"
	"net"
	"strconv"
	"time"
)

const(
	connType = "tcp"
)

func tcpRead(hostname string, port int, out chan []byte, errorChan chan error) {
	host := fmt.Sprintf("%s:%d", hostname, port)
	conn, err := net.DialTimeout(connType, host, time.Second * 5)
	if err != nil {
		errorChan <- err
		return
	}

	defer conn.Close()

	buf := make([]byte, 1024)
	for {
		_, err := conn.Read(buf)
		if err != nil {
			errorChan <- err
			return
		}
		out <- buf
	}
}

func writeSerial(device string, messages chan []byte, errorChan chan error) {
	c := &serial.Config{Name: device, Baud: 115200}
	s, err := serial.OpenPort(c)

	if err != nil {
		errorChan <- fmt.Errorf("error opening serial port: %v", err)
		return
	}

	for {
		msg := <- messages
		_, err := s.Write(msg)
		if err != nil {
			errorChan <- err
			return
		}
	}
}

func doRelay(hostname string, port int, device string) error {
	messages := make(chan []byte)
	errorChan := make(chan error)

	go tcpRead(hostname, port, messages, errorChan)
	go writeSerial(device, messages, errorChan)

	e := <- errorChan
	return e
}

func main() {
	app := app.New()

	w := app.NewWindow("RMonitor TCP Serial Relay")
	serialDevice := widget.NewEntry()
	serialDevice.PlaceHolder = "Serial device (e.g. COM1, /dev/tty.usbserial)"

	errLabel := widget.NewLabel("")

	hostname := widget.NewEntry()
	hostname.PlaceHolder = "Hostname (IP or FQDN)"

	port := widget.NewEntry()
	port.PlaceHolder = "Port Number"

	w.SetContent(widget.NewVBox(
		widget.NewLabel("RMonitor TCP Serial Relay"),
		serialDevice,
		widget.NewHBox(
			hostname,
			port),
		widget.NewHBox(
			widget.NewButton("Start", func() {
				go func() {
					errLabel.Text = "Streaming"
					errLabel.Refresh()
					if serialDevice.Text == "" {
						errLabel.Text = "Invalid serial device specified"
						errLabel.Refresh()
						return
					}

					if hostname.Text == "" {
						errLabel.Text = "Invalid hostname specified"
						errLabel.Refresh()
						return
					}

					if port.Text == "" {
						errLabel.Text = "Invalid port specified"
						errLabel.Refresh()
						return
					}

					portInt, err := strconv.Atoi(port.Text)
					if err != nil {
						errLabel.Text = "Could not convert port to int"
						errLabel.Refresh()
						return
					}

					err = doRelay(hostname.Text, portInt, serialDevice.Text)
					if err != nil {
						errLabel.Text = err.Error()
						errLabel.Refresh()
					}
				}()
			}),
			widget.NewButton("Quit", func() {
				app.Quit()
			}),
			),
		errLabel,
		))
	w.ShowAndRun()
}