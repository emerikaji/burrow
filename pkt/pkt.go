package pkt

import (
	"encoding/binary"
	"errors"
	"io"
	"strconv"
)

type Packet interface {
	Len() int
	Data() []byte
}

type Flush struct {
}

func (*Flush) Len() int {
	return 0
}

func (*Flush) Data() []byte {
	return nil
}

type DataPacket struct {
	data []byte
}

func (d *DataPacket) Len() int {
	return len(d.data)
}

func (d *DataPacket) Data() []byte {
	return d.data
}

// ─── Parse Pkt-line Data streams Into Flush Or Data Packets ──────────────────

func Parse(r io.Reader) (packets []Packet, err error) {
	for err == nil {
		pktLenBytes := make([]byte, 4)
		if err = binary.Read(r, binary.BigEndian, &pktLenBytes); err != nil {
			continue
		}

		var pktLen int64
		if pktLen, err = strconv.ParseInt(string(pktLenBytes), 16, 32); err != nil {
			continue
		}

		if pktLen == 0 {
			packets = append(packets, &Flush{})
			continue
		}

		packet := &DataPacket{
			data: make([]byte, pktLen-4),
		}

		if _, err = io.ReadFull(r, packet.data); err != nil {
			continue
		}

		packets = append(packets, packet)
	}

	if errors.Is(err, io.EOF) {
		err = nil
	}

	return
}

// ─────────────────────────────────────────────────────────────────────────────
