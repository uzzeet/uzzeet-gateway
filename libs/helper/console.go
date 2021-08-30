package helper

type Color int

type ColorType int

const (
	FOREGROUND ColorType = 1 + iota
)

const (
	DEFAULT Color = 1 + iota
	BLACK
	RED
	GREEN
	YELLOW
	BLUE
	MAGENTA
	CYAN
	LIGHT_GRAY
	DARK_GRAY
	LIGHT_RED
	LIGHT_GREEN
	LIGHT_YELLOW
	LIGHT_BLUE
	LIGHT_MAGENTA
	LIGHT_CYAN
	WHITE
)

var EscChar = "\x1B"

var ResetChar = EscChar + "[0m"

type ProgressBarOption struct {
	Size       float64 `json:"size"`
	Max        float64 `json:"max"`
	Value      float64 `json:"value"`
	FullColor  bool    `json:"full_color"`
	ValueColor Color
	BackColor  Color
}

func GetColorCode(c Color, t ColorType) (string, bool) {
	fcs := map[Color][]string{
		DEFAULT:       []string{"39", "49"},
		BLACK:         []string{"30", "40"},
		RED:           []string{"31", "41"},
		GREEN:         []string{"32", "42"},
		YELLOW:        []string{"33", "43"},
		BLUE:          []string{"34", "44"},
		MAGENTA:       []string{"35", "45"},
		CYAN:          []string{"36", "46"},
		LIGHT_GRAY:    []string{"37", "47"},
		DARK_GRAY:     []string{"90", "100"},
		LIGHT_RED:     []string{"91", "101"},
		LIGHT_GREEN:   []string{"92", "102"},
		LIGHT_YELLOW:  []string{"93", "103"},
		LIGHT_BLUE:    []string{"94", "104"},
		LIGHT_MAGENTA: []string{"95", "105"},
		LIGHT_CYAN:    []string{"96", "106"},
		WHITE:         []string{"97", "107"},
	}

	cc, ok := fcs[c]
	if ok {
		if t == FOREGROUND {
			return cc[0], true
		} else {
			return cc[1], true
		}
	}
	return "", false
}

func ApplyForeColor(s string, c Color) string {
	col, ok := GetColorCode(c, FOREGROUND)
	if ok {
		return EscChar + "[" + col + "m" + s + ResetChar
	}
	return s
}
