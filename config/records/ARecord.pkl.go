// Code generated from Pkl module `henrikvtcodes.tungsten.config.Records`. DO NOT EDIT.
package records

type ARecord interface {
	Record

	GetAddress() string
}

var _ ARecord = (*ARecordImpl)(nil)

type ARecordImpl struct {
	Address string `pkl:"address" json:"address"`

	Ttl uint32 `pkl:"ttl" json:"ttl"`
}

func (rcv *ARecordImpl) GetAddress() string {
	return rcv.Address
}

func (rcv *ARecordImpl) GetTtl() uint32 {
	return rcv.Ttl
}
