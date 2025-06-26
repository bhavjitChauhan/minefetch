package main

import (
	"errors"
	"fmt"
	"io"
)

// TODO: create tests

func readString(r io.Reader) (s string, err error) {
	x, err := readVarInt(r)
	if err != nil {
		return
	}

	buf := make([]byte, x)
	n, err := io.ReadFull(r, buf)
	if err != nil {
		return
	}
	if n != int(x) {
		err = errors.New(fmt.Sprint("expected", x, "bytes but got", n))
		return
	}

	s = string(buf)

	return
}

func writeString(w io.Writer, s string) error {
	err := writeVarInt(w, int32(len(s)))
	if err != nil {
		return err
	}

	_, err = w.Write([]byte(s))
	if err != nil {
		return err
	}

	return nil
}

// https://minecraft.wiki/w/Java_Edition_protocol/Data_types#VarInt_and_VarLong

const SEGMENT_BITS byte = 0b0111_1111
const CONTINUE_BIT byte = 0b1000_0000

func readVarInt(r io.Reader) (x int32, err error) {
	x = 0
	position := 0
	curr := make([]byte, 1)

	for {
		var n int
		n, err = r.Read(curr)
		if err != nil {
			return
		}
		if n != 1 {
			err = errors.New(fmt.Sprint("expected 1 byte but read ", n))
			return
		}

		x |= int32(curr[0]&SEGMENT_BITS) << position

		if curr[0]&CONTINUE_BIT == 0 {
			break
		}

		position += 7

		if position >= 32 {
			err = errors.New("VarInt is too big")
			return
		}
	}

	return x, nil
}

func writeVarInt(w io.Writer, x int32) error {
	uval := uint32(x)
	for {
		if (uval & ^uint32(SEGMENT_BITS)) == 0 {
			_, err := w.Write([]byte{byte(uval)})
			return err
		}

		_, err := w.Write([]byte{byte(uval&uint32(SEGMENT_BITS)) | CONTINUE_BIT})
		if err != nil {
			return err
		}
		uval >>= 7
	}
}
