package irm_net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"strings"

	hd "github.com/jguillaumes/go-hexdump"
	"github.com/jguillaumes/ims-injector/internal/irm"
	log "github.com/sirupsen/logrus"
)

func Do_interaction(host string, port uint16, irmTemplate irm.IRM, inc chan string, outc chan string, errc chan error) {

	sess, err := NewIMSconSess(host, port)
	if err != nil {
		errc <- fmt.Errorf("failed to create IMS connection session: %v", err)
		return
	}

	err = sess.Connect()
	if err != nil {
		errc <- fmt.Errorf("failed to connect to IMS: %v", err)
		return
	}
	defer sess.Close()

	sendBuffer := make([]byte, 0, 4*1024)  // Adjust buffer size as needed
	respBuffer := make([]byte, 0, 16*1024) // Adjust buffer size as needed

	for {
		log.Trace("Looping the loop")
		select {
		case msg := <-inc:
			if msg == "" {
				// If an empty message is received, exit silently (and end the goroutine)
				errc <- nil
				return
			}

			parts := strings.Split(msg, " ")
			trancode := parts[0]
			if len(trancode) > 8 {
				errc <- fmt.Errorf("transaction code %s is too long", trancode)
				continue
			}
			// Pad trancode to 8 bytes with spaces
			trancode = fmt.Sprintf("%-8s", trancode)

			irmTemplate.Irm_user.Irm_trncod = trancode

			log.Debug("Sending message to IMS: ", msg)
			len, err := prepareMessage(irmTemplate, msg, sendBuffer) // prepareMessage is a function that prepares the message for sending
			if err != nil {
				errc <- fmt.Errorf("failed to prepare message: %v", err)
				return
			}

			if log.IsLevelEnabled(log.DebugLevel) {
				d := hd.HexDump(sendBuffer[:len], "ISO8859-1")
				log.Debugf("Prepared message for IMS:\n%s", d)
			}

			// Send the message to IMS
			_, err = sess.conn.Write(sendBuffer[:len])
			if err != nil {
				errc <- fmt.Errorf("failed to send message to IMS: %v", err)
				return
			}

			// Read the response from IMS
			log.Debug("Waiting for response from IMS")
			n, err := sess.conn.Read(respBuffer)
			if err != nil {
				errc <- fmt.Errorf("failed to read response from IMS: %v", err)
				return
			}
			log.Debugf("Received response from IMS: %s", respBuffer[:n])

			outc <- string(respBuffer[:n]) // Convert response to string before sending

		case <-errc:
			return // Exit on error signal
		}
	}
}

func prepareMessage(irmTemplate irm.IRM, msg string, buf []byte) (int, error) {
	// Total length = Message length + IRM length + 4 bytes for the message llzz + 4 bytes for EOM
	if len(msg)+int(irmTemplate.Llll+8) > cap(buf) {
		return 0, fmt.Errorf("message too long for buffer. %d bytes required, %d bytes available", len(msg)+int(irmTemplate.Llll), cap(buf))
	}

	wbuff := bytes.NewBuffer(buf)

	// Make a copy of the IRM template to avoid modifying the original
	irm := irmTemplate
	// Set the length of the message in the IRM template
	irm.Llll = irmTemplate.Llll + uint32(len(msg))
	// Serialize the IRM into the buffer
	err := irm.Serialize(wbuff)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize IRM: %v", err)
	}
	// Prepare the message length and zz bytes
	msglen := len(msg)
	msglen_be := make([]byte, 2)
	binary.BigEndian.PutUint16(msglen_be, uint16(msglen))
	// Write the message length and zz bytes to the buffer
	wbuff.Write(msglen_be)
	wbuff.WriteByte(0) // zz byte, must be 0
	wbuff.WriteByte(0) // zz byte, must be 0

	// Copy the message into the buffer
	wbuff.WriteString(msg)

	// Add the EOM block
	wbuff.WriteByte(0)
	wbuff.WriteByte(0b00001000)
	wbuff.WriteByte(0)
	wbuff.WriteByte(0)

	return int(irmTemplate.Irm_len) + len(msg), nil
}
