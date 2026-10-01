package wallet

import (
	"errors"
	"fmt"
	"strings"
)

// CashAddr 编解码（https://github.com/bitcoincashorg/bitcoincash.org/blob/master/spec/cashaddr.md）。
// 只支持 160 位哈希的 P2PKH / P2SH，不支持 token-aware 地址（类型 2/3）。

const cashAddrCharset = "qpzry9x8gf2tvdw0s3jn54khce6mua7l"

const (
	CashAddrTypeP2PKH byte = 0
	CashAddrTypeP2SH  byte = 1
)

func cashPolymod(values []byte) uint64 {
	c := uint64(1)
	for _, d := range values {
		c0 := byte(c >> 35)
		c = ((c & 0x07ffffffff) << 5) ^ uint64(d)
		if c0&0x01 != 0 {
			c ^= 0x98f2bc8e61
		}
		if c0&0x02 != 0 {
			c ^= 0x79b76d99e2
		}
		if c0&0x04 != 0 {
			c ^= 0xf33e5fb3c4
		}
		if c0&0x08 != 0 {
			c ^= 0xae2eabe2a8
		}
		if c0&0x10 != 0 {
			c ^= 0x1e4f43e470
		}
	}
	return c ^ 1
}

// cashPrefixData 把前缀每个字符的低 5 位取出，再补一个 0。
func cashPrefixData(prefix string) []byte {
	out := make([]byte, 0, len(prefix)+1)
	for i := 0; i < len(prefix); i++ {
		out = append(out, prefix[i]&0x1f)
	}
	return append(out, 0)
}

func cashConvertBits(data []byte, from, to uint, pad bool) ([]byte, error) {
	var acc uint32
	var bits uint
	maxv := uint32(1)<<to - 1
	maxAcc := uint32(1)<<(from+to-1) - 1
	out := make([]byte, 0, len(data)*int(from)/int(to)+1)

	for _, b := range data {
		if uint32(b)>>from != 0 {
			return nil, errors.New("value out of range")
		}
		acc = (acc<<from | uint32(b)) & maxAcc
		bits += from
		for bits >= to {
			bits -= to
			out = append(out, byte(acc>>bits&maxv))
		}
	}
	if pad {
		if bits > 0 {
			out = append(out, byte(acc<<(to-bits)&maxv))
		}
	} else if bits >= from || acc<<(to-bits)&maxv != 0 {
		return nil, errors.New("invalid padding")
	}
	return out, nil
}

// EncodeCashAddr 返回带前缀的 CashAddr，例如 bitcoincash:qpm2qsznhks23z7629mms6s4cwef74vcwvy22gdx6a。
func EncodeCashAddr(prefix string, addrType byte, hash []byte) (string, error) {
	if len(hash) != 20 {
		return "", fmt.Errorf("cashaddr: hash must be 20 bytes, got %d", len(hash))
	}
	if addrType > 15 {
		return "", fmt.Errorf("cashaddr: invalid type %d", addrType)
	}
	prefix = strings.ToLower(prefix)

	// 版本字节: 高位保留(0) | 类型(4 位) | 大小码(3 位，160 位 = 0)
	data := append([]byte{addrType << 3}, hash...)
	payload, err := cashConvertBits(data, 8, 5, true)
	if err != nil {
		return "", err
	}

	checkInput := append(cashPrefixData(prefix), payload...)
	checkInput = append(checkInput, make([]byte, 8)...)
	cs := cashPolymod(checkInput)

	var sb strings.Builder
	sb.WriteString(prefix)
	sb.WriteByte(':')
	for _, d := range payload {
		sb.WriteByte(cashAddrCharset[d])
	}
	for i := 0; i < 8; i++ {
		sb.WriteByte(cashAddrCharset[(cs>>(5*uint(7-i)))&0x1f])
	}
	return sb.String(), nil
}

// DecodeCashAddr 解码 CashAddr。地址不带前缀时使用 defaultPrefix。
// 返回实际使用的前缀、地址类型 (0=P2PKH, 1=P2SH) 和 20 字节哈希。
func DecodeCashAddr(addr, defaultPrefix string) (prefix string, addrType byte, hash []byte, err error) {
	if addr != strings.ToLower(addr) && addr != strings.ToUpper(addr) {
		return "", 0, nil, errors.New("cashaddr: mixed case")
	}
	addr = strings.ToLower(addr)

	prefix, body := defaultPrefix, addr
	if i := strings.LastIndexByte(addr, ':'); i >= 0 {
		prefix, body = addr[:i], addr[i+1:]
	}
	if prefix == "" {
		return "", 0, nil, errors.New("cashaddr: missing prefix")
	}
	if len(body) <= 8 {
		return "", 0, nil, errors.New("cashaddr: too short")
	}

	data := make([]byte, len(body))
	for i := 0; i < len(body); i++ {
		idx := strings.IndexByte(cashAddrCharset, body[i])
		if idx < 0 {
			return "", 0, nil, fmt.Errorf("cashaddr: invalid character %q", body[i])
		}
		data[i] = byte(idx)
	}

	if cashPolymod(append(cashPrefixData(prefix), data...)) != 0 {
		return "", 0, nil, errors.New("cashaddr: bad checksum")
	}

	payload, err := cashConvertBits(data[:len(data)-8], 5, 8, false)
	if err != nil {
		return "", 0, nil, fmt.Errorf("cashaddr: %w", err)
	}
	if len(payload) != 21 {
		return "", 0, nil, fmt.Errorf("cashaddr: unsupported payload length %d (only 160-bit hashes)", len(payload)-1)
	}

	version := payload[0]
	if version&0x80 != 0 {
		return "", 0, nil, errors.New("cashaddr: reserved bit set")
	}
	if version&0x07 != 0 {
		return "", 0, nil, errors.New("cashaddr: unsupported hash size")
	}
	addrType = (version >> 3) & 0x0f
	if addrType > CashAddrTypeP2SH {
		return "", 0, nil, fmt.Errorf("cashaddr: unsupported address type %d (token-aware addresses are not supported)", addrType)
	}
	return prefix, addrType, payload[1:], nil
}
