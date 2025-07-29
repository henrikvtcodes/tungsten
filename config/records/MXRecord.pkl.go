// Code generated from Pkl module `henrikvtcodes.tungsten.config.Records`. DO NOT EDIT.
package records

type MXRecord interface {
	Record

	GetPreference() uint16

	GetExchange() any
}

var _ MXRecord = (*MXRecordImpl)(nil)

type MXRecordImpl struct {
	Preference uint16 `pkl:"preference" json:"preference"`

	Exchange any `pkl:"exchange" json:"exchange"`

	Ttl uint32 `pkl:"ttl" json:"ttl"`
}

func (rcv *MXRecordImpl) GetPreference() uint16 {
	return rcv.Preference
}

func (rcv *MXRecordImpl) GetExchange() any {
	return rcv.Exchange
}

func (rcv *MXRecordImpl) GetTtl() uint32 {
	return rcv.Ttl
}
