package irm_net

var IRM_messages = map[uint32]string{
	0x0004: "Exit request error message sent to client before socket termination. The socket is disconnected for IMS.",
	0x0008: "Error detected by IMS Connect and the socket is disconnected for IMS.",
	0x000C: "Error returned by IMS OTMA and the socket is disconnected for IMS.",
	0x0010: "Error returned by IMS OTMA when an OTMA sense code is returned in the \"Reason Code\" field of the RSM. The socket is disconnected for IMS.",
	0x0014: "Exit requests response message to HWSPWCH/PING request to be returned to client. IMS Connect keeps the socket connection because the PWCH/PING came in on a new socket connection or an existing persistent socket connection that is not in conversational mode or waiting for an ACK/NAK from the client application.",
	0x0018: "SCI error detected, see CSL codes for reason codes. The socket is disconnected for IMS.",
	0x001C: "OM error detected, see CSL codes for reason codes. The socket is disconnected for IMS.",
	0x0020: "IRM_TIMER value expired. When this return code is issued, the value of the corresponding reason code is not a code, but rather the time interval that was in effect for the IRM_TIMER. The socket is disconnected by IMS Connect.",
	0x0024: "A default IRM_TIMER value expired. Either the IRM_TIMER value specified was X'00' or an invalid value. When this return code is issued, the value of the corresponding reason code is not a code, but rather the time interval that was in effect for the IRM_TIMER. The socket is disconnected by IMS Connect.",
	0x0028: "IRM_TIMER value expired. When this return code is issued, the value of the corresponding reason code is not a code, but rather the time interval that was in effect for the IRM_TIMER. The connection is not disconnected. The socket remains connected.",
	0x002C: "The DATASTORE in no longer available.",
}
