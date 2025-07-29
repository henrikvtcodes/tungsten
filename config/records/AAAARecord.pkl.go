// Code generated from Pkl module `henrikvtcodes.tungsten.config.Records`. DO NOT EDIT.
package records

type AAAARecord interface {
	Record

	GetAddress() string
}

var _ AAAARecord = (*AAAARecordImpl)(nil)

type AAAARecordImpl struct {
	Address string `pkl:"address" json:"address"`

	Ttl uint32 `pkl:"ttl" json:"ttl"`
}

func (rcv *AAAARecordImpl) GetAddress() string {
	return rcv.Address
}

func (rcv *AAAARecordImpl) GetTtl() uint32 {
	return rcv.Ttl
}
