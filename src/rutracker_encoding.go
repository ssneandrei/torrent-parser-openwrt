package main

import "strings"

var cp1251Table = [...]rune{
 'Ђ', 'Ѓ', '‚', 'ѓ', '„', '…', '†', '‡',
 '€', '‰', 'Љ', '‹', 'Њ', 'Ќ', 'Ћ', 'Џ',
 'ђ', '‘', '’', '“', '”', '•', '–', '—',
 '\uFFFD', '™', 'љ', '›', 'њ', 'ќ', 'ћ', 'џ',
 '\u00A0', 'Ў', 'ў', 'Ј', '¤', 'Ґ', '¦', '§',
 'Ё', '©', 'Є', '«', '¬', '\u00AD', '®', 'Ї',
 '°', '±', 'І', 'і', 'ґ', 'µ', '¶', '·',
 'ё', '№', 'є', '»', 'ј', 'Ѕ', 'ѕ', 'ї',
}

func decodeRuTrackerText(data []byte) string {
 var out strings.Builder
 out.Grow(len(data) * 2)

 for _, c:= range data {
 switch {
 case c < 0x80:
 out.WriteByte(c)

 case c >= 0xC0:
 out.WriteRune(rune(0x0410 + int(c-0xC0)))

 default:
 out.WriteRune(cp1251Table[int(c)-0x80])
 }
 }

 return out.String()
}
