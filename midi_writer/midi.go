package main

import (
	"flag"
	"fmt"
	"math"
	"os"
	"strconv"
	"strings"
)

var note_map = map[string]uint8{
	"c":  0,
	"c+": 1,
	"d-": 1,
	"d":  2,
	"d+": 3,
	"e-": 3,
	"e":  4,
	"f":  5,
	"f+": 6,
	"g-": 6,
	"g":  7,
	"g+": 8,
	"a-": 8,
	"a":  9,
	"a+": 10,
	"b-": 10,
	"b":  11,
}

const (
	Timebase = 480

	NoteOn     = 0x90
	NoteOff    = 0x80
	ProgChange = 0xc0

	Inst_piano  = 0
	Inst_organ  = 19
	Inst_guitar = 25
	Inst_flute  = 73
)

var (
	Filename    string
	Tempo       uint32 = 60
	Numerator   byte   = 6
	Denominator uint   = 8
	Instrument  byte   = Inst_piano
)

type Note struct {
	num     []uint8
	time    uint32
	vel     uint8
	channel uint8
}

type MidiData []byte

func (midi *MidiData) push(vals ...byte) {
	*midi = append(*midi, vals...)
}
func (midi *MidiData) push_str(vals string) {
	*midi = append(*midi, []byte(vals)...)
}
func (midi *MidiData) push_u16(v uint16) {
	midi.push(byte(v>>8&0xff), byte(v>>0&0xff))
}
func (midi *MidiData) push_u24(v uint32) {
	midi.push(byte(v>>16&0xff), byte(v>>8&0xff), byte(v>>0&0xff))
}
func (midi *MidiData) push_u32(v uint32) {
	midi.push(byte(v>>24&0xff), byte(v>>16&0xff), byte(v>>8&0xff), byte(v>>0&0xff))
}
func (midi *MidiData) push_delta(v uint32) {
	if v > 0x0fffffff {
		panic("delta must be <= 0x0fffffff")
	}
	var buf uint32 = v & 0x7f
	for (v >> 7) > 0 {
		v >>= 7
		buf <<= 8
		buf |= 0x80
		buf += v & 0x7f
	}
	for {
		midi.push(byte(buf & 0xff))
		if buf&0x80 == 0 {
			break
		}
		buf >>= 8
	}
}

func generate_header(midi_format, n_tracks, timebase uint16) (res MidiData) {
	res.push_str("MThd")
	const header_len = 6
	res.push_u32(header_len)
	res.push_u16(midi_format)
	res.push_u16(n_tracks)
	res.push_u16(timebase)
	return res
}

func log2[T byte | uint](n T) T {
	i := T(0)
	for ; n > 1; i++ {
		n >>= 1
	}
	return i
}
func min[T byte | uint](a T, b ...T) T {
	for _, v := range b {
		if a > v {
			a = v
		}
	}
	return a
}

func generate_track(track []Note) (res MidiData) {
	timefrac_set := []byte{0xff, 0x58, 0x04}
	res.push_delta(0)
	res.push(timefrac_set...)
	res.push(Numerator,
		byte(log2(Denominator)),
		96/byte(min(Denominator, 32)), // clock/beat
		32/4)                          // 32nd/quarter

	tempo_set := []byte{0xff, 0x51, 0x03}
	res.push_delta(0)
	res.push(tempo_set...)
	res.push_u24(60 * 1000000 / (Tempo)) // usec/beat

	res.push_delta(0)
	res.push(ProgChange + track[0].channel)
	res.push(Instrument)

	for _, n := range track {
		if !(n.channel <= 15 && n.vel <= 127 && len(n.num) > 0) {
			panic(fmt.Sprintf("invalid note: %v", n))
		}
		for _, v := range n.num {
			res.push_delta(0)
			res.push(NoteOn + n.channel)
			res.push(v)
			res.push(n.vel)
		}
		for i, v := range n.num {
			if i == 0 {
				res.push_delta(n.time)
			} else {
				res.push_delta(0)
			}
			res.push(NoteOff + n.channel)
			res.push(v)
			res.push(0)
		}
	}
	end_of_track := []byte{0xff, 0x2f, 0x00}
	res.push_delta(0)
	res.push(end_of_track...)
	return res
}

func count_left(s string, sub string) int {
	n := 0
	for ; len(s) >= len(sub) && s[:len(sub)] == sub; n++ {
		s = s[len(sub):]
	}
	return n
}

func parse_length(time *uint32, tok string) uint32 {
	var length uint32
	t := tok
	if n, _ := fmt.Sscanf(tok, "%v%s", &length, &t); n >= 1 && length == 0 {
		return 0
	}
	if length != 0 {
		*time = Timebase * 4 / length
	}

	if dots := count_left(t, "."); dots > 0 {
		*time += *time * ((1 << dots) - 1) / (1 << dots)
		t = t[dots:]
	}

	ties := uint32(count_left(t, "^"))
	t = t[ties:]

	return *time * (1 + ties)
}

func get_notename(tok string) (notenum uint8, rest string, ok bool) {
	for i := len(tok); i >= 1; i-- {
		if notenum, ok := note_map[tok[:i]]; ok {
			return notenum, tok[i:], true
		}
	}
	return 0, tok, false
}

func parse_track(channel uint8, tokens []string) (notes []Note) {
	type t_ctx struct {
		time     uint32
		vel      uint8
		octave   uint8
		is_chord bool
	}

	const central_C = 60
	ctx := t_ctx{
		time:     Timebase,
		vel:      80,
		octave:   central_C,
		is_chord: false,
	}

	for _, tok := range tokens {
		if len(tok) < 1 {
			continue
		}
		if notenum, tok, ok := get_notename(tok); ok {
			duration := parse_length(&ctx.time, tok)
			if !ctx.is_chord {
				n := Note{num: []uint8{}, time: duration, vel: ctx.vel, channel: channel}
				notes = append(notes, n)
			}
			n := &notes[len(notes)-1]
			n.num = append(n.num, ctx.octave+notenum)
			n.time = duration
		} else if tok[:1] == "`" {
			if !ctx.is_chord {
				n := Note{num: []uint8{}, time: ctx.time, vel: ctx.vel, channel: channel}
				notes = append(notes, n)
			} else if ctx.is_chord {
				duration := parse_length(&ctx.time, tok[1:])
				n := &notes[len(notes)-1]
				n.time = duration
			}
			ctx.is_chord = !ctx.is_chord
		} else if tok[:1] == "r" {
			duration := parse_length(&ctx.time, tok[1:])
			n := Note{num: []uint8{0}, time: duration, vel: 0, channel: channel}
			notes = append(notes, n)
		} else if tok[:1] == "l" {
			ctx.time = parse_length(&ctx.time, tok[1:])
		} else if tok[:1] == "o" {
			oct, _ := strconv.ParseUint(tok[1:], 10, 32)
			ctx.octave = 12 * uint8(oct)
		} else if tok == ">" {
			ctx.octave += 12
		} else if tok == "<" {
			ctx.octave -= 12
		} else {
			println("unsupported token: " + tok)
		}
	}
	fmt.Printf("tok:%v notes: %v\n", tokens, notes)
	return notes
}

const (
	silent_night_chorus = "` < < b- > b- > d f `8. ` < < b- > b- > e- g `16 ` < < b- > b- > d f `8 ` < < b- > f b- > d `4. " +
		"` < < b- > b- > d f `8. ` < < b- > b- > e- g `16 ` < < b- > b- > d f `8 ` < < b- > f b- > d `4. " +
		"` < f a > e- > c < `4 ` < f a > e- > c < `8 ` < f > c e- a `4. " +
		"` < < b- > b- > d b- `4 ` < < b- > f > d b- `8 ` < < b- > b- > d f `4. " +
		"` < e- b- > e- g `4 ` < e- b- > e- g `8 ` < e- g > g b- `8. ` < e- a > f a `16 ` < e- b- > e- g `8 " +
		"` < < b- > b- > d f `8. ` < < b- > b- > e- g `16 ` < < b- > b- > d f `8 ` < < b- > f b- > d `4. " +
		"` < e- e- b- > g `4 ` < e- b- > e- g `8 ` < e- g > g b- `8. ` < e- a > f a `16 ` < e- b- > e- g `8 " +
		"` < < b- > b- > d f `8. ` < < b- > b- > e- g `16 ` < < b- > b- > d f `8 ` < < b- > f b- > d `4. " +
		"` < f a > e- > c < `4 ` < f a > e- > c < `8 ` < f > c e- > e- < `8. ` < f a > e- > c < `16 ` < f > c e- a `8 " +
		"` < < b- > b- > d b- `4. ` < < b- > b- > f > d < ` " +
		"` < < b- > f > d b- `8 ` < < b- > f > d f ` ` < < b- > f b- > d ` ` < < f > f > d f `8. ` < < f > f > c e- `16 ` < < f > e- a > c `8 " +
		"` < < b- > d b- b- `2"

	silent_night_guitar = "l8 < < b- > f f l16 < b- > f b- > d < b- f " +
		"l8 < b- > f f l16 < b- > f b- > d < b- f " +
		"l8 < f > f f l16 < f > f a > c f < f " +
		"l8 < b- > f f l16 < b- > f b- > d < b- f " +
		"l8 < b- > e- e- l8 < b- > e- e- " +
		"l8 < b- > f f l16 < b- > f b- > d < b- f " +
		"l8 < b- > e- e- l8 < b- > e- e- " +
		"l8 < b- > f f l16 < b- > f b- > d < b- f " +
		"l8 < f > f f l8 < f > f f " +
		"l16 < b- > f b- > d < b- f l16 < b- > f b- > d < b- f " +
		"l8 < b- > f f l8 < f > f f " +
		"l8 < b- f d < b-^ "
)

func set_flags() {
	flag.StringVar(&Filename, "o", "out.mid", "output file")
	tempo_ := flag.Uint("t", 60, "tempo (quarters/min)")
	numerator_ := flag.Uint("n", 6, "numerator of time")
	denominator_ := flag.Uint("d", 8, "denominator of time (must be a pow of 2)")
	instrument_ := flag.Uint("i", Inst_guitar, "midi instrument number (0-127)")
	flag.Parse()

	if !(0 < *tempo_ && *tempo_ <= math.MaxInt16) {
		panic("tempo must be > 0")
	}
	if *numerator_ == 0 {
		panic("numerator must be > 0")
	}
	if *denominator_&(*denominator_-1) != 0 {
		panic("denominator must be a pow of 2")
	}
	if !(*instrument_ <= 127) {
		panic("instrument number must be 0 .. 127")
	}
	Tempo = uint32(*tempo_)
	Numerator = byte(*numerator_)
	Denominator = uint(*denominator_)
	Instrument = byte(*instrument_)
}

func main() {
	set_flags()

	tracks := []string{silent_night_chorus, silent_night_guitar}
	if flag.NArg() > 0 {
		tracks = flag.Args()
	}
	res := MidiData{}

	const midi_format = 1
	header := generate_header(midi_format, uint16(len(tracks)), Timebase)
	res.push(header...)

	for i, t := range tracks {
		notes := parse_track(uint8(i), strings.Split(t, " "))
		block := generate_track(notes)
		res.push_str("MTrk")
		res.push_u32(uint32(len(block)))
		res.push(block...)
	}

	file, _ := os.Create(Filename)
	defer file.Close()
	file.Write(res)
}
