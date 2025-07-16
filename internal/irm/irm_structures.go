package irm

import "fmt"

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
		Irm_timer:       30,              // Default timer value = 30 seconds
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
		Irm_f3:           IRM_F3_SYNCCONF,
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
func (u *IRM_USER) Serialize(buf []byte) error {
	if len(buf) < 76 {
		return fmt.Errorf("buffer too small for IRM_USER serialization")
	}

	buf[0] = u.Irm_f1
	buf[1] = u.Irm_f2
	buf[2] = u.Irm_f3
	buf[3] = u.Irm_f4
	copy(buf[4:12], u.Irm_trncod)
	copy(buf[12:20], u.Irm_imsdestid)
	copy(buf[20:28], u.Irm_lterm)
	copy(buf[28:36], u.Irm_racf_userid)
	copy(buf[36:44], u.Irm_racf_grpname)
	copy(buf[44:52], u.Irm_racf_pw)
	copy(buf[52:60], u.Irm_appl_nm)
	copy(buf[60:68], u.Irm_rerout_nm)
	copy(buf[68:76], u.Irm_rt_altcid)

	return nil
}

// Serialize IRM into a provided byte slice, checking for sufficient length
// The numbers must be serialized in big-endian order
func (irm *IRM) Serialize(buf []byte) error {
	if len(buf) < 108 { // 32 + 76 = 108 bytes
		return fmt.Errorf("buffer too small for IRM serialization")
	}

	buf[0] = byte(irm.Llll >> 24)
	buf[1] = byte(irm.Llll >> 16)
	buf[2] = byte(irm.Llll >> 8)
	buf[3] = byte(irm.Llll)

	buf[4] = byte(irm.Irm_len >> 8)
	buf[5] = byte(irm.Irm_len)

	buf[6] = irm.Irm_arch
	buf[7] = irm.Irm_f0

	copy(buf[8:16], irm.Irm_id)

	buf[16] = byte(irm.Irm_nak_rsncode >> 8)
	buf[17] = byte(irm.Irm_nak_rsncode)

	buf[18] = byte(irm.irm_res1 >> 8)
	buf[19] = byte(irm.irm_res1)

	buf[20] = irm.Irm_f5
	buf[21] = irm.Irm_timer
	buf[22] = irm.Irm_soct
	buf[23] = irm.Irm_es

	copy(buf[24:32], irm.Irm_clientid)

	return irm.Irm_user.Serialize(buf[32:108])
}
