package irm_net

import (
	"bytes"
	"encoding/binary"
	"fmt"
	"io"
	"strings"

	hd "github.com/jguillaumes/go-hexdump"
	"github.com/jguillaumes/ims-injector/internal/irm"
	log "github.com/sirupsen/logrus"
)

// Do_interaction interacts with IMS connect to send transactions and receive results.
//
// This function is intended to be run as a goroutine. It will read the transactions from the
// inc channel and write the responses to the outc channel. In case of error, it will be reported using
// the errc channel.
// num is a number representing the goroutine, and it is used to build an unique clientId if necessary.
// irmTemplate contains the common information used to interact with IMS Connect. host and port are
// self explanatory.
func Do_interaction(num int, host string, port uint16, irmTemplate irm.IRM, inc chan string, outc chan string, errc chan error) {

	var clientId string
	if num > 0 {
		baseClientId := strings.TrimSpace(irmTemplate.Irm_clientid)
		clientId = fmt.Sprintf("%s%d", baseClientId, num)
	} else {
		clientId = irmTemplate.Irm_clientid
	}

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

	log.Debugf("Concurrent interaction processor %d with clientid %s started.", num, clientId)

	sendBuffer := make([]byte, 0, 4*1024) // Adjust buffer size as needed
	respBuffer := make([]byte, 256*1024)  // Adjust buffer size as needed

	for {
		msg, ok := <-inc
		if !ok {
			// Check for closed channel
			errc <- nil // Signal end of goroutine
			break
		}

		parts := strings.Split(msg, " ")
		trancode := parts[0]
		if len(trancode) > 8 {
			errc <- fmt.Errorf("transaction code %s is too long", trancode)
			continue
		}
		// Pad trancode to 8 bytes with spaces
		trancode = fmt.Sprintf("%-8s", trancode)

		// Make a local copy of irmTemplate
		irm := irmTemplate
		irm.Irm_clientid = clientId
		irm.Irm_user.Irm_trncod = trancode

		log.Debug("Sending message to IMS: ", msg)
		len, err := prepareMessage(&irm, msg, sendBuffer) // prepareMessage is a function that prepares the message for sending
		if err != nil {
			errc <- fmt.Errorf("failed to prepare message: %v", err)
			break
		}

		if log.IsLevelEnabled(log.TraceLevel) {
			d := hd.HexDump(sendBuffer[:len], "ISO8859-1")
			log.Debugf("Prepared message for IMS:\n%s", d)
		}

		// Send the message to IMS
		n, err := sess.conn.Write(sendBuffer[:len])
		if err != nil {
			errc <- fmt.Errorf("failed to send message to IMS: %v", err)
			break // Unexpected condition, end process
		}
		log.Debugf("Wrote %d tx bytes.\n", n)

		// Read the response from IMS
		log.Debug("Waiting for response from IMS")
		n, err = io.ReadAtLeast(sess.conn, respBuffer, 4)
		if err != nil && err != io.EOF {
			errc <- fmt.Errorf("failed to read response from IMS: %v", err)
			break // Unexpected condition, end process
		}
		llll := binary.BigEndian.Uint32(respBuffer[:4])
		if int(llll) > n {
			n, err = io.ReadAtLeast(sess.conn, respBuffer[n:], int(llll)-n)
			if err != nil && err != io.EOF {
				errc <- fmt.Errorf("failed to read response from IMS: %v", err)
				break // Unexpected condition, end process
			}
		}
		log.Debugf("Read %d tx response bytes.\n", n)

		response, need_ack, nowait_ack, resperr := analyzeResponse(respBuffer)

		if need_ack {
			log.Debug("ACK was requested")
			// Send ack
			err = send_ack(sess, &irmTemplate, nowait_ack, sendBuffer, respBuffer)
			if err != nil {
				errc <- fmt.Errorf("failed to read response from IMS ACK: %v", err)
				break // Unexpected condition, end process
			}
		}
		if resperr != nil {
			log.Warnf("Error received from IMS Connect: %v\n", resperr)
			continue // Skip this transaction and continue
		}

		fullresp := strings.Join(response, "\n")
		log.Tracef("Response:\n%s\n", fullresp)
		outc <- fullresp

	}
	log.Debugf("Concurrent interaction processor %d ended.", num)
}

// send_ack prepares and sends an ACK message to IMS Connect
// If the nowait flag is specified it will use the IRM_NO_WAIT value for the IRM timeout
// and will *not* wait for a response. Otherwise, it will perform a read after sending
// the ACK.
func send_ack(sess *IMSconSess, irmTemplate *irm.IRM, nowait bool, sendBuffer []byte, respBuffer []byte) error {
	irm_ack := *irmTemplate
	irm_ack.Llll += 4 // EOM
	irm_ack.Irm_user.Irm_f4 = irm.IRM_F4_ACK
	if nowait {
		irm_ack.Irm_timer = 0xE9 // IRM no wait
		irm_ack.Irm_user.Irm_f1 |= irm.IRM_F1_NOWAIT
	} else {
		irm_ack.Irm_timer = 0x1E // 0.5 seconds
	}
	wbuff := bytes.NewBuffer(sendBuffer)
	err := irm_ack.Serialize(wbuff)
	if err != nil {
		return err
	}
	// Add the EOM block
	wbuff.WriteByte(0)
	wbuff.WriteByte(0b00000100)
	wbuff.WriteByte(0)
	wbuff.WriteByte(0)

	log.Debug("Sending ack to IMS: ")
	n, err := sess.conn.Write(sendBuffer[:irm_ack.Llll])
	if err != nil {
		return err
	}
	log.Debugf("Wrote %d ack bytes.\n", n)

	if !nowait {
		n, err = io.ReadAtLeast(sess.conn, respBuffer, 4)
		if err != nil {
			return err
		}
		llll := binary.BigEndian.Uint32(respBuffer[:4])
		if n < int(llll) {
			_, err = io.ReadAtLeast(sess.conn, respBuffer[4:], int(llll)-n)
		}
		log.Debugf("Read %d ack response bytes.\n", n)
		if log.IsLevelEnabled(log.TraceLevel) {
			d := hd.HexDump(respBuffer[:llll], "ISO8859-1")
			log.Tracef("Response to ACK:\n%s", d)
		}
	}
	return err
}

// prepareMessage prepares a message to be sent to IMS Connect.
// The message is built serializing the irm block and adding the segment corresponding
// to the transaction text specified by msg. The message to be sent is
// built in the buf byte slice.
func prepareMessage(irm *irm.IRM, msg string, buf []byte) (int, error) {
	// Total length = Message length + IRM length + 4 bytes for the message llzz + 4 bytes for EOM
	if len(msg)+int(irm.Llll+8) > cap(buf) {
		return 0, fmt.Errorf("message too long for buffer. %d bytes required, %d bytes available", len(msg)+int(irm.Llll), cap(buf))
	}

	wbuff := bytes.NewBuffer(buf)

	// Set the length of the message in the IRM template
	irm.Llll = irm.Llll + uint32(len(msg)+8)
	// Serialize the IRM into the buffer
	err := irm.Serialize(wbuff)
	if err != nil {
		return 0, fmt.Errorf("failed to serialize IRM: %v", err)
	}
	// Prepare the message length and zz bytes
	msglen := len(msg) + 4
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
	wbuff.WriteByte(0b00000100)
	wbuff.WriteByte(0)
	wbuff.WriteByte(0)

	return wbuff.Len(), nil
}

// analyzeResponse parses an IMS Connect response buffer
// If the buffer corresponds to a transaction response, it builds a slice of strings,
// one element for response segment. Notice the results are undefined if the response
// contains non-text elements.
// It also checks the different status blocks to determine if an ACK is required, and
// if the NOWAIT function is available.
func analyzeResponse(buffer []byte) ([]string, bool, bool, error) {
	var ackRequired = false
	var ackNowait = false
	var err error = nil
	var response = make([]string, 0, 100)

	bufReader := bytes.NewBuffer(buffer)

	remaining := int(binary.BigEndian.Uint32(bufReader.Next(4))) // Total message len (per HWSSMPL1)
	remaining -= 4
	for remaining > 0 {
		if bufReader.Len() < 4 {
			log.Errorf("inconsistency: not enough bytes to proceed in buffer. Bytes in buffer=%d, expected=%d", bufReader.Len(), remaining)
			if bufReader.Available() > 0 {
				log.Error(hd.HexDump(bufReader.AvailableBuffer(), "ISO8859-1"))
			}
		}
		seglen := binary.BigEndian.Uint16(bufReader.Next(2))   // Segment length
		segFlags := binary.BigEndian.Uint16(bufReader.Next(2)) // Segment flags
		segData := bufReader.Next(int(seglen - 4))             // Rest of segment data
		remaining -= int(seglen)
		if remaining < 0 {
			log.Warnf("remaining bytes went negative (%d)", remaining)
		}

		// Check for possible control data
		if seglen >= 12 {
			identifier_bytes := segData[:8]
			identifier := string(identifier_bytes)
			switch identifier {
			case "*REQMOD*":
				{
					// MODNAME present in transaction response. Read it, log it and ignore
					modName_bytes := segData[8:16]
					modName := string(modName_bytes)
					log.Infof("Modname present in response: %-8s", modName)
					continue
				}
			case "*REQSTS*":
				{
					// Error/status response
					if (segFlags & 0x2000) != 0 {
						// ACK required
						ackRequired = true
					}
					if segFlags&0x0002 != 0 {
						// ACK Nowait supported
						ackNowait = true
					}
					rsm_retcode := binary.BigEndian.Uint32(segData[8:12])
					rsm_rsncode := binary.BigEndian.Uint32(segData[12:16])
					errmsg, ok := IRM_messages[rsm_retcode]
					if !ok {
						errmsg = "No text available"
					}
					var errrsn string
					switch rsm_rsncode {
					case 0x0010:
						errrsn = fmt.Sprintf("OTMA reason code %04X", rsm_rsncode)
					case 0x0018, 0x001C:
						errrsn = fmt.Sprintf("CSL reason code %04X", rsm_rsncode)
					default:
						errrsn, ok = IRM_reasons[rsm_rsncode]
						if !ok {
							errrsn = "No text available"
						}
					}
					err = fmt.Errorf("error returned by IMS Connect: %s: %s (RC=%04X, RSN=%04X)", errmsg, errrsn, rsm_retcode, rsm_rsncode)
					continue
				}

			case "*CSMOKY*":
				{
					// End of message
					if (segFlags & 0x2000) != 0 {
						// ACK required
						ackRequired = true
					}
					if segFlags&0x0002 != 0 {
						// ACK Nowait supported
						ackNowait = true
					}
					continue // End the loop
				}
			default:
				// Actual transaction response data
				response_line := string(segData)
				response = append(response, response_line)
				log.Tracef("Response line received: %s", response_line)
				continue
			}
		} else {
			// Actual transaction response data
			response_line := string(segData)
			response = append(response, response_line)
			log.Tracef("Response line received: %s", response_line)
			continue
		}
	}
	if remaining != 0 {
		log.Warnf("%d spurious characters detected!", remaining)
	}
	return response, ackRequired, ackNowait, err
}
