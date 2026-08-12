package main

import (
 "net/url"
 "sort"
 "strings"
)

func encodeCP1251Bytes(s string) []byte {
 out:= make([]byte, 0, len(s))

 for _, r:= range s {
 switch {
 case r < 0x80:
 out = append(out, byte(r))

 case r >= 0x0410 && r <= 0x044F:
 out = append(out, byte(0xC0+int(r-0x0410)))

 default:
 encoded:= byte('?')

 for i, candidate:= range cp1251Table {
 if candidate == r && candidate!= '\uFFFD' {
 encoded = byte(0x80 + i)
 break
 }
 }

 out = append(out, encoded)
 }
 }

 return out
}

func percentEncodeBytes(data []byte) string {
 const hex = "0123456789ABCDEF"

 var out strings.Builder
 out.Grow(len(data) * 3)

 for _, c:= range data {
 switch {
 case c >= 'a' && c <= 'z',
 c >= 'A' && c <= 'Z',
 c >= '0' && c <= '9',
 c == '-', c == '_', c == '.', c == '~':
 out.WriteByte(c)

 case c == ' ':
 out.WriteByte('+')

 default:
 out.WriteByte('%')
 out.WriteByte(hex[c>>4])
 out.WriteByte(hex[c&15])
 }
 }

 return out.String()
}

func formEncodeCP1251(values url.Values) string {
 keys:= make([]string, 0, len(values))
 for key:= range values {
 keys = append(keys, key)
 }
 sort.Strings(keys)

 parts:= make([]string, 0)

 for _, key:= range keys {
 encodedKey:= percentEncodeBytes([]byte(key))

 for _, value:= range values[key] {
 parts = append(
 parts,
 encodedKey+"="+percentEncodeBytes(encodeCP1251Bytes(value)),
 )
 }
 }

 return strings.Join(parts, "&")
}
