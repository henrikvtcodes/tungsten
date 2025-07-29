// Code generated from Pkl module `henrikvtcodes.tungsten.config.Records`. DO NOT EDIT.
package records

type TXTRecord interface {
	Record

	GetContent() string
}

var _ TXTRecord = (*TXTRecordImpl)(nil)

type TXTRecordImpl struct {
	Content string `pkl:"content" json:"content"`

	Ttl uint32 `pkl:"ttl" json:"ttl"`
}

func (rcv *TXTRecordImpl) GetContent() string {
	return rcv.Content
}

func (rcv *TXTRecordImpl) GetTtl() uint32 {
	return rcv.Ttl
}
