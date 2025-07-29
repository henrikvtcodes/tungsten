// Code generated from Pkl module `henrikvtcodes.tungsten.config.Records`. DO NOT EDIT.
package records

type CNAMERecord interface {
	Record

	GetTarget() string
}

var _ CNAMERecord = (*CNAMERecordImpl)(nil)

type CNAMERecordImpl struct {
	Target string `pkl:"target" json:"target"`

	Ttl uint32 `pkl:"ttl" json:"ttl"`
}

func (rcv *CNAMERecordImpl) GetTarget() string {
	return rcv.Target
}

func (rcv *CNAMERecordImpl) GetTtl() uint32 {
	return rcv.Ttl
}
