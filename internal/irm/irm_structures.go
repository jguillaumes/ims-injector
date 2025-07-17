package irm

import (
	"bytes"
	"encoding/binary"
	"fmt"
)

// +
// IRM_COMMON: 28 bytes + 4 bytes for the total length
// The length field includes itself
// -
type IRM struct {
	Llll            uint32
	Irm_len         uint16
	Irm_arch        uint8
	Irm_f0          uint8
	Irm_id          string
	Irm_nak_rsncode uint16
	irm_res1        uint16
	Irm_f5          uint8
	Irm_timer       uint8
	Irm_soct        uint8
	Irm_es          uint8
	Irm_clientid    string
	Irm_user        IRM_USER
}

// +
// ARCH 0x01 user header.
// 76 bytes of length
// -
type IRM_USER struct {
	Irm_f1           uint8
	Irm_f2           uint8
	Irm_f3           uint8
	Irm_f4           uint8
	Irm_trncod       string
	Irm_imsdestid    string
	Irm_lterm        string
	Irm_racf_userid  string
	Irm_racf_grpname string
	Irm_racf_pw      string
	Irm_appl_nm      string
	Irm_rerout_nm    string
	Irm_rt_altcid    string
}

func NewIRM() *IRM {
	return &IRM{
		Llll:            32 + 76,       // Total length of the IRM_COMMON structure
		Irm_len:         28 + 76,       // Length of the IRM structure
		Irm_arch:        IRM_ARCH_LVL1, // Architecture level 1
		Irm_id:          "*SAMPLE*",
		Irm_nak_rsncode: 0,
		irm_res1:        0,
		Irm_f5:          0,
		Irm_timer:       10,              // Default timer value = 10 seconds
		Irm_soct:        SOCT_PERSISTENT, // Default socket type = Persistent
		Irm_es:          0,               // NO Unicode used
		Irm_clientid:    "        ",      // Let the EXIT assign the client ID
		Irm_user:        *NewIRM_USER(),
	}
}

func NewIRM_USER() *IRM_USER {
	return &IRM_USER{
		Irm_f1:           IRM_F1_TRNEXP,
		Irm_f2:           IRM_F2_CM1 | IRM_F2_GENCLID,
		Irm_f3:           IRM_F3_SYNCNONE,
		Irm_f4:           IRM_F4_SENDREC,
		Irm_trncod:       "        ",
		Irm_imsdestid:    "        ",
		Irm_lterm:        "        ",
		Irm_racf_userid:  "        ",
		Irm_racf_grpname: "        ",
		Irm_racf_pw:      "        ",
		Irm_appl_nm:      "        ",
		Irm_rerout_nm:    "        ",
		Irm_rt_altcid:    "        ",
	}
}

// Serialize IRM_USER into a provided byte slice, checking for sufficient length
func (u *IRM_USER) Serialize(buf *bytes.Buffer) error {
	if buf.Available() < 76 {
		return fmt.Errorf("buffer too small for IRM_USER serialization. %d bytes required, %d bytes provided", 76, buf.Available())
	}

	buf.WriteByte(u.Irm_f1)
	buf.WriteByte(u.Irm_f2)
	buf.WriteByte(u.Irm_f3)
	buf.WriteByte(u.Irm_f4)
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_trncod))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_imsdestid))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_lterm))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_racf_userid))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_racf_grpname))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_racf_pw))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_appl_nm))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_rerout_nm))
	buf.WriteString(fmt.Sprintf("%-8s", u.Irm_rt_altcid))

	return nil
}

// Serialize IRM into a provided byte buffer, checking for sufficient length
// The numbers must be serialized in big-endian order
func (irm *IRM) Serialize(buf *bytes.Buffer) error {
	if buf.Available() < 108 { // 32 + 76 = 108 bytes
		return fmt.Errorf("buffer too small for IRM serialization. %d bytes required, %d bytes provided", 108, buf.Available())
	}

	// Set the length of the IRM structure
	llll_be := make([]byte, 4)
	binary.BigEndian.PutUint32(llll_be, irm.Llll)
	buf.Write(llll_be)

	ll_be := make([]byte, 2)
	binary.BigEndian.PutUint16(ll_be, irm.Irm_len)
	buf.Write(ll_be)

	buf.WriteByte(irm.Irm_arch)
	buf.WriteByte(irm.Irm_f0)
	buf.WriteString(fmt.Sprintf("%-8s", irm.Irm_id))

	buf.WriteByte(0) // nak_rsncode high byte
	buf.WriteByte(0) // nak_rsncode low byte

	buf.WriteByte(0) // irm_res1 high byte
	buf.WriteByte(0) // irm_res1 low byte

	buf.WriteByte(irm.Irm_f5)
	buf.WriteByte(irm.Irm_timer)
	buf.WriteByte(irm.Irm_soct)
	buf.WriteByte(irm.Irm_es)

	buf.WriteString(fmt.Sprintf("%-8s", irm.Irm_clientid))

	return irm.Irm_user.Serialize(buf)
}
