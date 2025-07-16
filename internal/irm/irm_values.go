package irm

// Arch levels for IRM structures
const (
	ARCH_LVL0 = 0x00 // Architecture level 0
	ARCH_LVL1 = 0x01 // Architecture level 1
)

// Values for F0
const (
	IRM_F0_SYNONLY = 0x80
	IRM_F0_SYNASIN = 0x40
	IRM_F0_SYNCNAK = 0x20
	IRM_F0_NAKRSN  = 0x10
	IRM_F0_EXENS   = 0x04
	IRM_F0_XMLTD   = 0x01
	IRM_F0_XML_D   = 0x02
)

// Values for socket type
const (
	SOCT_TRANSACTION   = 0x00 // Transaction socket
	SOCT_PERSISTENT    = 0x10 // Persistent socket
	SOCT_NONPERSISTENT = 0x40 // Non-persistent socket
)

// Values for IRM_F1
const (
	IRM_F1_MFSREQ = 0x80 // MFS MOD requested
	IRM_F1_CIDREQ = 0x40 // Client ID requested
	IRM_F1_UC     = 0x20 // Unicode message
	IRM_F1_UCTC   = 0x10 // Unicode transaction code
	IRM_F1_SOARSP = 0x04 // No message text in ACKs for send-only with ACK requests
	IRM_F1_NOWAIT = 0x02 // Send-and-receive CM0 with NOWAIT option
	IRM_F1_TRNEXP = 0x01 // Set transaction expiration time
	IRM_F1_NOMOD  = 0x00 // No MFS MOD requested
)

// Values for IRM_F2
const (
	IRM_F2_CM0      = 0x40 // CM0 message (commit-then-send)
	IRM_F2_CM1      = 0x20 // CM1 message (send-then-commit)
	IRM_F2_SENDALTP = 0x02
	IRM_F2_GENCLID  = 0x01 // Generate unique client ID
)

// Vaues for IRM_F3
const (
	IRM_F3_SYNCNONE = 0x00 // Sync level = None
	IRM_F3_SYNCCONF = 0x01 // Sync level = Confirm
	IRM_F3_SYNCPTX  = 0x02 // Sync level = Syncpt
	IRM_F3_PURGE    = 0x04 // Purge undeliverable CM0 output messages
	IRM_F3_REROUT   = 0x08 // Reroute undeliverable CM0 output messages
	IRM_F3_ORDER    = 0x10 // Set ordered delivery of send-only messages
	IRM_F3_IGNPURGE = 0x20 // Ignore PURG calls in multisegment CM0 output
	IRM_F3_DFS2082  = 0x40 // Receive DFS2082 if a CM0 interaction does not answer
	IRM_F3_CANCID   = 0x80 // Cancel duplicate client ID
)

// Values for IRM_F4 (message type)
const (
	IRM_F4_ACK      = uint8('A') // Acknowledgment message
	IRM_F4_CANTIMER = uint8('C') // Cancel timer message
	IRM_F4_DEALLOC  = uint8('D') // Deallocate conversation
	IRM_F4_SNDONLYE = uint8('J') // Send-only with possible IMSCON error
	IRM_F4_SNDONLYA = uint8('K') // Send-only with ACK
	IRM_F4_SYNRESPA = uint8('L') // Synchronous callout requiring ACK
	IRM_F4_SYNRESP  = uint8('M') // Synchronous callout
	IRM_F4_NACK     = uint8('N') // NAK message
	IRM_F4_RESUMET  = uint8('R') // Resume TPIPE request
	IRM_F4_SENDONLY = uint8('S') // Send-only message
	IRM_F4_SENDREC  = uint8(' ') // Send-and-receive message
)
