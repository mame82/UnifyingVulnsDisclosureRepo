package hid

import "fmt"

type HIDKey byte
type HIDMod byte

var (
	StringToUsbKey    = generateStr2Key()
	StringToUsbModKey = generateStr2Mod()
)

const (
	HID_KEY_RESERVED           HIDKey = 0x00
	HID_KEY_ERROR_ROLLOVER     HIDKey = 0x01
	HID_KEY_POST_FAIL          HIDKey = 0x02
	HID_KEY_ERROR_UNDEFINED    HIDKey = 0x03
	HID_KEY_A                  HIDKey = 0x04
	HID_KEY_B                  HIDKey = 0x05
	HID_KEY_C                  HIDKey = 0x06
	HID_KEY_D                  HIDKey = 0x07 // Keyboard d and D
	HID_KEY_E                  HIDKey = 0x08 // Keyboard e and E
	HID_KEY_F                  HIDKey = 0x09 // Keyboard f and F
	HID_KEY_G                  HIDKey = 0x0a // Keyboard g and G
	HID_KEY_H                  HIDKey = 0x0b // Keyboard h and H
	HID_KEY_I                  HIDKey = 0x0c // Keyboard i and I
	HID_KEY_J                  HIDKey = 0x0d // Keyboard j and J
	HID_KEY_K                  HIDKey = 0x0e // Keyboard k and K
	HID_KEY_L                  HIDKey = 0x0f // Keyboard l and L
	HID_KEY_M                  HIDKey = 0x10 // Keyboard m and M
	HID_KEY_N                  HIDKey = 0x11 // Keyboard n and N
	HID_KEY_O                  HIDKey = 0x12 // Keyboard o and O
	HID_KEY_P                  HIDKey = 0x13 // Keyboard p and P
	HID_KEY_Q                  HIDKey = 0x14 // Keyboard q and Q
	HID_KEY_R                  HIDKey = 0x15 // Keyboard r and R
	HID_KEY_S                  HIDKey = 0x16 // Keyboard s and S
	HID_KEY_T                  HIDKey = 0x17 // Keyboard t and T
	HID_KEY_U                  HIDKey = 0x18 // Keyboard u and U
	HID_KEY_V                  HIDKey = 0x19 // Keyboard v and V
	HID_KEY_W                  HIDKey = 0x1a // Keyboard w and W
	HID_KEY_X                  HIDKey = 0x1b // Keyboard x and X
	HID_KEY_Y                  HIDKey = 0x1c // Keyboard y and Y
	HID_KEY_Z                  HIDKey = 0x1d // Keyboard z and Z
	HID_KEY_1                  HIDKey = 0x1e // Keyboard 1 and !
	HID_KEY_2                  HIDKey = 0x1f // Keyboard 2 and @
	HID_KEY_3                  HIDKey = 0x20 // Keyboard 3 and #
	HID_KEY_4                  HIDKey = 0x21 // Keyboard 4 and $
	HID_KEY_5                  HIDKey = 0x22 // Keyboard 5 and %
	HID_KEY_6                  HIDKey = 0x23 // Keyboard 6 and ^
	HID_KEY_7                  HIDKey = 0x24 // Keyboard 7 and &
	HID_KEY_8                  HIDKey = 0x25 // Keyboard 8 and *
	HID_KEY_9                  HIDKey = 0x26 // Keyboard 9 and (
	HID_KEY_0                  HIDKey = 0x27 // Keyboard 0 and )
	HID_KEY_ENTER              HIDKey = 0x28 // Keyboard Return (ENTER)
	HID_KEY_ESC                HIDKey = 0x29 // Keyboard ESCAPE
	HID_KEY_BACKSPACE          HIDKey = 0x2a // Keyboard DELETE (Backspace)
	HID_KEY_TAB                HIDKey = 0x2b // Keyboard Tab
	HID_KEY_SPACE              HIDKey = 0x2c // Keyboard Spacebar
	HID_KEY_MINUS              HIDKey = 0x2d // Keyboard - and _
	HID_KEY_EQUAL              HIDKey = 0x2e // Keyboard = and +
	HID_KEY_LEFTBRACE          HIDKey = 0x2f // Keyboard [ and {
	HID_KEY_RIGHTBRACE         HIDKey = 0x30 // Keyboard ] and }
	HID_KEY_BACKSLASH          HIDKey = 0x31 // Keyboard \ and |
	HID_KEY_HASHTILDE          HIDKey = 0x32 // Keyboard Non-US # and ~
	HID_KEY_SEMICOLON          HIDKey = 0x33 // Keyboard ; and :
	HID_KEY_APOSTROPHE         HIDKey = 0x34 // Keyboard ' and "
	HID_KEY_GRAVE              HIDKey = 0x35 // Keyboard ` and ~
	HID_KEY_COMMA              HIDKey = 0x36 // Keyboard , and <
	HID_KEY_DOT                HIDKey = 0x37 // Keyboard . and >
	HID_KEY_SLASH              HIDKey = 0x38 // Keyboard / and ?
	HID_KEY_CAPSLOCK           HIDKey = 0x39 // Keyboard Caps Lock
	HID_KEY_F1                 HIDKey = 0x3a // Keyboard F1
	HID_KEY_F2                 HIDKey = 0x3b // Keyboard F2
	HID_KEY_F3                 HIDKey = 0x3c // Keyboard F3
	HID_KEY_F4                 HIDKey = 0x3d // Keyboard F4
	HID_KEY_F5                 HIDKey = 0x3e // Keyboard F5
	HID_KEY_F6                 HIDKey = 0x3f // Keyboard F6
	HID_KEY_F7                 HIDKey = 0x40 // Keyboard F7
	HID_KEY_F8                 HIDKey = 0x41 // Keyboard F8
	HID_KEY_F9                 HIDKey = 0x42 // Keyboard F9
	HID_KEY_F10                HIDKey = 0x43 // Keyboard F10
	HID_KEY_F11                HIDKey = 0x44 // Keyboard F11
	HID_KEY_F12                HIDKey = 0x45 // Keyboard F12
	HID_KEY_SYSRQ              HIDKey = 0x46 // Keyboard Print Screen
	HID_KEY_SCROLLLOCK         HIDKey = 0x47 // Keyboard Scroll Lock
	HID_KEY_PAUSE              HIDKey = 0x48 // Keyboard Pause
	HID_KEY_INSERT             HIDKey = 0x49 // Keyboard Insert
	HID_KEY_HOME               HIDKey = 0x4a // Keyboard Home
	HID_KEY_PAGEUP             HIDKey = 0x4b // Keyboard Page Up
	HID_KEY_DELETE             HIDKey = 0x4c // Keyboard Delete Forward
	HID_KEY_END                HIDKey = 0x4d // Keyboard End
	HID_KEY_PAGEDOWN           HIDKey = 0x4e // Keyboard Page Down
	HID_KEY_RIGHT              HIDKey = 0x4f // Keyboard Right Arrow
	HID_KEY_LEFT               HIDKey = 0x50 // Keyboard Left Arrow
	HID_KEY_DOWN               HIDKey = 0x51 // Keyboard Down Arrow
	HID_KEY_UP                 HIDKey = 0x52 // Keyboard Up Arrow
	HID_KEY_NUMLOCK            HIDKey = 0x53 // Keyboard Num Lock and Clear
	HID_KEY_KPSLASH            HIDKey = 0x54 // Keypad /
	HID_KEY_KPASTERISK         HIDKey = 0x55 // Keypad *
	HID_KEY_KPMINUS            HIDKey = 0x56 // Keypad -
	HID_KEY_KPPLUS             HIDKey = 0x57 // Keypad +
	HID_KEY_KPENTER            HIDKey = 0x58 // Keypad ENTER
	HID_KEY_KP1                HIDKey = 0x59 // Keypad 1 and End
	HID_KEY_KP2                HIDKey = 0x5a // Keypad 2 and Down Arrow
	HID_KEY_KP3                HIDKey = 0x5b // Keypad 3 and PageDn
	HID_KEY_KP4                HIDKey = 0x5c // Keypad 4 and Left Arrow
	HID_KEY_KP5                HIDKey = 0x5d // Keypad 5
	HID_KEY_KP6                HIDKey = 0x5e // Keypad 6 and Right Arrow
	HID_KEY_KP7                HIDKey = 0x5f // Keypad 7 and Home
	HID_KEY_KP8                HIDKey = 0x60 // Keypad 8 and Up Arrow
	HID_KEY_KP9                HIDKey = 0x61 // Keypad 9 and Page Up
	HID_KEY_KP0                HIDKey = 0x62 // Keypad 0 and Insert
	HID_KEY_KPDOT              HIDKey = 0x63 // Keypad . and Delete
	HID_KEY_102ND              HIDKey = 0x64 // Keyboard Non-US \ and |
	HID_KEY_COMPOSE            HIDKey = 0x65 // Keyboard Application
	HID_KEY_POWER              HIDKey = 0x66 // Keyboard Power
	HID_KEY_KPEQUAL            HIDKey = 0x67 // Keypad =
	HID_KEY_F13                HIDKey = 0x68 // Keyboard F13
	HID_KEY_F14                HIDKey = 0x69 // Keyboard F14
	HID_KEY_F15                HIDKey = 0x6a // Keyboard F15
	HID_KEY_F16                HIDKey = 0x6b // Keyboard F16
	HID_KEY_F17                HIDKey = 0x6c // Keyboard F17
	HID_KEY_F18                HIDKey = 0x6d // Keyboard F18
	HID_KEY_F19                HIDKey = 0x6e // Keyboard F19
	HID_KEY_F20                HIDKey = 0x6f // Keyboard F20
	HID_KEY_F21                HIDKey = 0x70 // Keyboard F21
	HID_KEY_F22                HIDKey = 0x71 // Keyboard F22
	HID_KEY_F23                HIDKey = 0x72 // Keyboard F23
	HID_KEY_F24                HIDKey = 0x73 // Keyboard F24
	HID_KEY_OPEN               HIDKey = 0x74 // Keyboard Execute
	HID_KEY_HELP               HIDKey = 0x75 // Keyboard Help
	HID_KEY_PROPS              HIDKey = 0x76 // Keyboard Menu
	HID_KEY_FRONT              HIDKey = 0x77 // Keyboard Select
	HID_KEY_STOP               HIDKey = 0x78 // Keyboard Stop
	HID_KEY_AGAIN              HIDKey = 0x79 // Keyboard Again
	HID_KEY_UNDO               HIDKey = 0x7a // Keyboard Undo
	HID_KEY_CUT                HIDKey = 0x7b // Keyboard Cut
	HID_KEY_COPY               HIDKey = 0x7c // Keyboard Copy
	HID_KEY_PASTE              HIDKey = 0x7d // Keyboard Paste
	HID_KEY_FIND               HIDKey = 0x7e // Keyboard Find
	HID_KEY_MUTE               HIDKey = 0x7f // Keyboard Mute
	HID_KEY_VOLUMEUP           HIDKey = 0x80 // Keyboard Volume Up
	HID_KEY_VOLUMEDOWN         HIDKey = 0x81 // Keyboard Volume Down
	HID_KEY_KPCOMMA            HIDKey = 0x85 // Keypad Comma
	HID_KEY_RO                 HIDKey = 0x87 // Keyboard International1
	HID_KEY_KATAKANAHIRAGANA   HIDKey = 0x88 // Keyboard International2
	HID_KEY_YEN                HIDKey = 0x89 // Keyboard International3
	HID_KEY_HENKAN             HIDKey = 0x8a // Keyboard International4
	HID_KEY_MUHENKAN           HIDKey = 0x8b // Keyboard International5
	HID_KEY_KPJPCOMMA          HIDKey = 0x8c // Keyboard International6
	HID_KEY_HANGEUL            HIDKey = 0x90 // Keyboard LANG1
	HID_KEY_HANJA              HIDKey = 0x91 // Keyboard LANG2
	HID_KEY_KATAKANA           HIDKey = 0x92 // Keyboard LANG3
	HID_KEY_HIRAGANA           HIDKey = 0x93 // Keyboard LANG4
	HID_KEY_ZENKAKUHANKAKU     HIDKey = 0x94 // Keyboard LANG5
	HID_KEY_KPLEFTPAREN        HIDKey = 0xb6 // Keypad (
	HID_KEY_KPRIGHTPAREN       HIDKey = 0xb7 // Keypad )
	HID_KEY_LEFTCTRL           HIDKey = 0xe0 // Keyboard Left Control
	HID_KEY_LEFTSHIFT          HIDKey = 0xe1 // Keyboard Left Shift
	HID_KEY_LEFTALT            HIDKey = 0xe2 // Keyboard Left Alt
	HID_KEY_LEFTMETA           HIDKey = 0xe3 // Keyboard Left GUI
	HID_KEY_RIGHTCTRL          HIDKey = 0xe4 // Keyboard Right Control
	HID_KEY_RIGHTSHIFT         HIDKey = 0xe5 // Keyboard Right Shift
	HID_KEY_RIGHTALT           HIDKey = 0xe6 // Keyboard Right Alt
	HID_KEY_RIGHTMETA          HIDKey = 0xe7 // Keyboard Right GUI
	HID_KEY_MEDIA_PLAYPAUSE    HIDKey = 0xe8
	HID_KEY_MEDIA_STOPCD       HIDKey = 0xe9
	HID_KEY_MEDIA_PREVIOUSSONG HIDKey = 0xea
	HID_KEY_MEDIA_NEXTSONG     HIDKey = 0xeb
	HID_KEY_MEDIA_EJECTCD      HIDKey = 0xec
	HID_KEY_MEDIA_VOLUMEUP     HIDKey = 0xed
	HID_KEY_MEDIA_VOLUMEDOWN   HIDKey = 0xee
	HID_KEY_MEDIA_MUTE         HIDKey = 0xef
	HID_KEY_MEDIA_WWW          HIDKey = 0xf0
	HID_KEY_MEDIA_BACK         HIDKey = 0xf1
	HID_KEY_MEDIA_FORWARD      HIDKey = 0xf2
	HID_KEY_MEDIA_STOP         HIDKey = 0xf3
	HID_KEY_MEDIA_FIND         HIDKey = 0xf4
	HID_KEY_MEDIA_SCROLLUP     HIDKey = 0xf5
	HID_KEY_MEDIA_SCROLLDOWN   HIDKey = 0xf6
	HID_KEY_MEDIA_EDIT         HIDKey = 0xf7
	HID_KEY_MEDIA_SLEEP        HIDKey = 0xf8
	HID_KEY_MEDIA_COFFEE       HIDKey = 0xf9
	HID_KEY_MEDIA_REFRESH      HIDKey = 0xfa
	HID_KEY_MEDIA_CALC         HIDKey = 0xfb
)

const (
	HID_MOD_KEY_LEFT_CONTROL  HIDMod = 0x01
	HID_MOD_KEY_LEFT_SHIFT    HIDMod = 0x02
	HID_MOD_KEY_LEFT_ALT      HIDMod = 0x04
	HID_MOD_KEY_LEFT_GUI      HIDMod = 0x08
	HID_MOD_KEY_RIGHT_CONTROL HIDMod = 0x10
	HID_MOD_KEY_RIGHT_SHIFT   HIDMod = 0x20
	HID_MOD_KEY_RIGHT_ALT     HIDMod = 0x40
	HID_MOD_KEY_RIGHT_GUI     HIDMod = 0x80
)

func (c HIDMod) String() string {
	switch c {
	case HID_MOD_KEY_LEFT_CONTROL:
		return "MOD_LEFT_CONTROL"
	case HID_MOD_KEY_LEFT_SHIFT:
		return "MOD_LEFT_SHIFT"
	case HID_MOD_KEY_LEFT_ALT:
		return "MOD_LEFT_ALT"
	case HID_MOD_KEY_LEFT_GUI:
		return "MOD_LEFT_GUI"
	case HID_MOD_KEY_RIGHT_CONTROL:
		return "MOD_RIGHT_CONTROL"
	case HID_MOD_KEY_RIGHT_SHIFT:
		return "MOD_RIGHT_SHIFT"
	case HID_MOD_KEY_RIGHT_ALT:
		return "MOD_RIGHT_ALT"
	case HID_MOD_KEY_RIGHT_GUI:
		return "MOD_RIGHT_GUI"
	default:
		return fmt.Sprintf("unknown HID code %#02x", c)
	}
}

func (c HIDKey) String() string {
	switch c {
	case HID_KEY_RESERVED:
		return "KEY_RESERVED"
	case HID_KEY_ERROR_ROLLOVER:
		return "KEY_ERROR_ROLLOVER"
	case HID_KEY_POST_FAIL:
		return "KEY_POST_FAIL"
	case HID_KEY_ERROR_UNDEFINED:
		return "KEY_ERROR_UNDEFINED"
	case HID_KEY_A:
		return "KEY_A"
	case HID_KEY_B:
		return "KEY_B"
	case HID_KEY_C:
		return "KEY_C"
	case HID_KEY_D:
		return "KEY_D"
	case HID_KEY_E:
		return "KEY_E"
	case HID_KEY_F:
		return "KEY_F"
	case HID_KEY_G:
		return "KEY_G"
	case HID_KEY_H:
		return "KEY_H"
	case HID_KEY_I:
		return "KEY_I"
	case HID_KEY_J:
		return "KEY_J"
	case HID_KEY_K:
		return "KEY_K"
	case HID_KEY_L:
		return "KEY_L"
	case HID_KEY_M:
		return "KEY_M"
	case HID_KEY_N:
		return "KEY_N"
	case HID_KEY_O:
		return "KEY_O"
	case HID_KEY_P:
		return "KEY_P"
	case HID_KEY_Q:
		return "KEY_Q"
	case HID_KEY_R:
		return "KEY_R"
	case HID_KEY_S:
		return "KEY_S"
	case HID_KEY_T:
		return "KEY_T"
	case HID_KEY_U:
		return "KEY_U"
	case HID_KEY_V:
		return "KEY_V"
	case HID_KEY_W:
		return "KEY_W"
	case HID_KEY_X:
		return "KEY_X"
	case HID_KEY_Y:
		return "KEY_Y"
	case HID_KEY_Z:
		return "KEY_Z"
	case HID_KEY_1:
		return "KEY_1"
	case HID_KEY_2:
		return "KEY_2"
	case HID_KEY_3:
		return "KEY_3"
	case HID_KEY_4:
		return "KEY_4"
	case HID_KEY_5:
		return "KEY_5"
	case HID_KEY_6:
		return "KEY_6"
	case HID_KEY_7:
		return "KEY_7"
	case HID_KEY_8:
		return "KEY_8"
	case HID_KEY_9:
		return "KEY_9"
	case HID_KEY_0:
		return "KEY_0"
	case HID_KEY_ENTER:
		return "KEY_ENTER"
	case HID_KEY_ESC:
		return "KEY_ESC"
	case HID_KEY_BACKSPACE:
		return "KEY_BACKSPACE"
	case HID_KEY_TAB:
		return "KEY_TAB"
	case HID_KEY_SPACE:
		return "KEY_SPACE"
	case HID_KEY_MINUS:
		return "KEY_MINUS"
	case HID_KEY_EQUAL:
		return "KEY_EQUAL"
	case HID_KEY_LEFTBRACE:
		return "KEY_LEFTBRACE"
	case HID_KEY_RIGHTBRACE:
		return "KEY_RIGHTBRACE"
	case HID_KEY_BACKSLASH:
		return "KEY_BACKSLASH"
	case HID_KEY_HASHTILDE:
		return "KEY_HASHTILDE"
	case HID_KEY_SEMICOLON:
		return "KEY_SEMICOLON"
	case HID_KEY_APOSTROPHE:
		return "KEY_APOSTROPHE"
	case HID_KEY_GRAVE:
		return "KEY_GRAVE"
	case HID_KEY_COMMA:
		return "KEY_COMMA"
	case HID_KEY_DOT:
		return "KEY_DOT"
	case HID_KEY_SLASH:
		return "KEY_SLASH"
	case HID_KEY_CAPSLOCK:
		return "KEY_CAPSLOCK"
	case HID_KEY_F1:
		return "KEY_F1"
	case HID_KEY_F2:
		return "KEY_F2"
	case HID_KEY_F3:
		return "KEY_F3"
	case HID_KEY_F4:
		return "KEY_F4"
	case HID_KEY_F5:
		return "KEY_F5"
	case HID_KEY_F6:
		return "KEY_F6"
	case HID_KEY_F7:
		return "KEY_F7"
	case HID_KEY_F8:
		return "KEY_F8"
	case HID_KEY_F9:
		return "KEY_F9"
	case HID_KEY_F10:
		return "KEY_F10"
	case HID_KEY_F11:
		return "KEY_F11"
	case HID_KEY_F12:
		return "KEY_F12"
	case HID_KEY_SYSRQ:
		return "KEY_SYSRQ"
	case HID_KEY_SCROLLLOCK:
		return "KEY_SCROLLLOCK"
	case HID_KEY_PAUSE:
		return "KEY_PAUSE"
	case HID_KEY_INSERT:
		return "KEY_INSERT"
	case HID_KEY_HOME:
		return "KEY_HOME"
	case HID_KEY_PAGEUP:
		return "KEY_PAGEUP"
	case HID_KEY_DELETE:
		return "KEY_DELETE"
	case HID_KEY_END:
		return "KEY_END"
	case HID_KEY_PAGEDOWN:
		return "KEY_PAGEDOWN"
	case HID_KEY_RIGHT:
		return "KEY_RIGHT"
	case HID_KEY_LEFT:
		return "KEY_LEFT"
	case HID_KEY_DOWN:
		return "KEY_DOWN"
	case HID_KEY_UP:
		return "KEY_UP"
	case HID_KEY_NUMLOCK:
		return "KEY_NUMLOCK"
	case HID_KEY_KPSLASH:
		return "KEY_KPSLASH"
	case HID_KEY_KPASTERISK:
		return "KEY_KPASTERISK"
	case HID_KEY_KPMINUS:
		return "KEY_KPMINUS"
	case HID_KEY_KPPLUS:
		return "KEY_KPPLUS"
	case HID_KEY_KPENTER:
		return "KEY_KPENTER"
	case HID_KEY_KP1:
		return "KEY_KP1"
	case HID_KEY_KP2:
		return "KEY_KP2"
	case HID_KEY_KP3:
		return "KEY_KP3"
	case HID_KEY_KP4:
		return "KEY_KP4"
	case HID_KEY_KP5:
		return "KEY_KP5"
	case HID_KEY_KP6:
		return "KEY_KP6"
	case HID_KEY_KP7:
		return "KEY_KP7"
	case HID_KEY_KP8:
		return "KEY_KP8"
	case HID_KEY_KP9:
		return "KEY_KP9"
	case HID_KEY_KP0:
		return "KEY_KP0"
	case HID_KEY_KPDOT:
		return "KEY_KPDOT"
	case HID_KEY_102ND:
		return "KEY_102ND"
	case HID_KEY_COMPOSE:
		return "KEY_COMPOSE"
	case HID_KEY_POWER:
		return "KEY_POWER"
	case HID_KEY_KPEQUAL:
		return "KEY_KPEQUAL"
	case HID_KEY_F13:
		return "KEY_F13"
	case HID_KEY_F14:
		return "KEY_F14"
	case HID_KEY_F15:
		return "KEY_F15"
	case HID_KEY_F16:
		return "KEY_F16"
	case HID_KEY_F17:
		return "KEY_F17"
	case HID_KEY_F18:
		return "KEY_F18"
	case HID_KEY_F19:
		return "KEY_F19"
	case HID_KEY_F20:
		return "KEY_F20"
	case HID_KEY_F21:
		return "KEY_F21"
	case HID_KEY_F22:
		return "KEY_F22"
	case HID_KEY_F23:
		return "KEY_F23"
	case HID_KEY_F24:
		return "KEY_F24"
	case HID_KEY_OPEN:
		return "KEY_OPEN"
	case HID_KEY_HELP:
		return "KEY_HELP"
	case HID_KEY_PROPS:
		return "KEY_PROPS"
	case HID_KEY_FRONT:
		return "KEY_FRONT"
	case HID_KEY_STOP:
		return "KEY_STOP"
	case HID_KEY_AGAIN:
		return "KEY_AGAIN"
	case HID_KEY_UNDO:
		return "KEY_UNDO"
	case HID_KEY_CUT:
		return "KEY_CUT"
	case HID_KEY_COPY:
		return "KEY_COPY"
	case HID_KEY_PASTE:
		return "KEY_PASTE"
	case HID_KEY_FIND:
		return "KEY_FIND"
	case HID_KEY_MUTE:
		return "KEY_MUTE"
	case HID_KEY_VOLUMEUP:
		return "KEY_VOLUMEUP"
	case HID_KEY_VOLUMEDOWN:
		return "KEY_VOLUMEDOWN"
	case HID_KEY_KPCOMMA:
		return "KEY_KPCOMMA"
	case HID_KEY_RO:
		return "KEY_RO"
	case HID_KEY_KATAKANAHIRAGANA:
		return "KEY_KATAKANAHIRAGANA"
	case HID_KEY_YEN:
		return "KEY_YEN"
	case HID_KEY_HENKAN:
		return "KEY_HENKAN"
	case HID_KEY_MUHENKAN:
		return "KEY_MUHENKAN"
	case HID_KEY_KPJPCOMMA:
		return "KEY_KPJPCOMMA"
	case HID_KEY_HANGEUL:
		return "KEY_HANGEUL"
	case HID_KEY_HANJA:
		return "KEY_HANJA"
	case HID_KEY_KATAKANA:
		return "KEY_KATAKANA"
	case HID_KEY_HIRAGANA:
		return "KEY_HIRAGANA"
	case HID_KEY_ZENKAKUHANKAKU:
		return "KEY_ZENKAKUHANKAKU"
	case HID_KEY_KPLEFTPAREN:
		return "KEY_KPLEFTPAREN"
	case HID_KEY_KPRIGHTPAREN:
		return "KEY_KPRIGHTPAREN"
	case HID_KEY_LEFTCTRL:
		return "KEY_LEFTCTRL"
	case HID_KEY_LEFTSHIFT:
		return "KEY_LEFTSHIFT"
	case HID_KEY_LEFTALT:
		return "KEY_LEFTALT"
	case HID_KEY_LEFTMETA:
		return "KEY_LEFTMETA"
	case HID_KEY_RIGHTCTRL:
		return "KEY_RIGHTCTRL"
	case HID_KEY_RIGHTSHIFT:
		return "KEY_RIGHTSHIFT"
	case HID_KEY_RIGHTALT:
		return "KEY_RIGHTALT"
	case HID_KEY_RIGHTMETA:
		return "KEY_RIGHTMETA"
	case HID_KEY_MEDIA_PLAYPAUSE:
		return "KEY_MEDIA_PLAYPAUSE"
	case HID_KEY_MEDIA_STOPCD:
		return "KEY_MEDIA_STOPCD"
	case HID_KEY_MEDIA_PREVIOUSSONG:
		return "KEY_MEDIA_PREVIOUSSONG"
	case HID_KEY_MEDIA_NEXTSONG:
		return "KEY_MEDIA_NEXTSONG"
	case HID_KEY_MEDIA_EJECTCD:
		return "KEY_MEDIA_EJECTCD"
	case HID_KEY_MEDIA_VOLUMEUP:
		return "KEY_MEDIA_VOLUMEUP"
	case HID_KEY_MEDIA_VOLUMEDOWN:
		return "KEY_MEDIA_VOLUMEDOWN"
	case HID_KEY_MEDIA_MUTE:
		return "KEY_MEDIA_MUTE"
	case HID_KEY_MEDIA_WWW:
		return "KEY_MEDIA_WWW"
	case HID_KEY_MEDIA_BACK:
		return "KEY_MEDIA_BACK"
	case HID_KEY_MEDIA_FORWARD:
		return "KEY_MEDIA_FORWARD"
	case HID_KEY_MEDIA_STOP:
		return "KEY_MEDIA_STOP"
	case HID_KEY_MEDIA_FIND:
		return "KEY_MEDIA_FIND"
	case HID_KEY_MEDIA_SCROLLUP:
		return "KEY_MEDIA_SCROLLUP"
	case HID_KEY_MEDIA_SCROLLDOWN:
		return "KEY_MEDIA_SCROLLDOWN"
	case HID_KEY_MEDIA_EDIT:
		return "KEY_MEDIA_EDIT"
	case HID_KEY_MEDIA_SLEEP:
		return "KEY_MEDIA_SLEEP"
	case HID_KEY_MEDIA_COFFEE:
		return "KEY_MEDIA_COFFEE"
	case HID_KEY_MEDIA_REFRESH:
		return "KEY_MEDIA_REFRESH"
	case HID_KEY_MEDIA_CALC:
		return "KEY_MEDIA_CALC"
	default:
		return fmt.Sprintf("UNKNOWN_HID_CODE_%02X", byte(c))
	}
}

func generateStr2Mod() (s2m map[string]HIDMod) {
	s2m = make(map[string]HIDMod)
	s2m["MOD_LEFT_CONTROL"] = HID_MOD_KEY_LEFT_CONTROL
	s2m["MOD_LEFT_SHIFT"] = HID_MOD_KEY_LEFT_SHIFT
	s2m["MOD_LEFT_ALT"] = HID_MOD_KEY_LEFT_ALT
	s2m["MOD_LEFT_GUI"] = HID_MOD_KEY_LEFT_GUI
	s2m["MOD_RIGHT_CONTROL"] = HID_MOD_KEY_RIGHT_CONTROL
	s2m["MOD_RIGHT_SHIFT"] = HID_MOD_KEY_RIGHT_SHIFT
	s2m["MOD_RIGHT_ALT"] = HID_MOD_KEY_RIGHT_ALT
	s2m["MOD_RIGHT_GUI"] = HID_MOD_KEY_RIGHT_GUI
	return
}

func generateStr2Key() (s2k map[string]HIDKey) {
	s2k = make(map[string]HIDKey)
	s2k["KEY_RESERVED"] = HID_KEY_RESERVED
	s2k["KEY_ERROR_ROLLOVER"] = HID_KEY_ERROR_ROLLOVER
	s2k["KEY_POST_FAIL"] = HID_KEY_POST_FAIL
	s2k["KEY_ERROR_UNDEFINED"] = HID_KEY_ERROR_UNDEFINED
	s2k["KEY_A"] = HID_KEY_A // Keyboard a and A
	s2k["KEY_B"] = HID_KEY_B // Keyboard b and B
	s2k["KEY_C"] = HID_KEY_C // Keyboard c and C
	s2k["KEY_D"] = HID_KEY_D // Keyboard d and D
	s2k["KEY_E"] = HID_KEY_E // Keyboard e and E
	s2k["KEY_F"] = HID_KEY_F // Keyboard f and F
	s2k["KEY_G"] = HID_KEY_G // Keyboard g and G
	s2k["KEY_H"] = HID_KEY_H // Keyboard h and H
	s2k["KEY_I"] = HID_KEY_I // Keyboard i and I
	s2k["KEY_J"] = HID_KEY_J //0x0d // Keyboard j and J
	s2k["KEY_K"] = HID_KEY_K //0x0e // Keyboard k and K
	s2k["KEY_L"] = HID_KEY_L //0x0f // Keyboard l and L
	s2k["KEY_M"] = HID_KEY_M //0x10 // Keyboard m and M
	s2k["KEY_N"] = HID_KEY_N //0x11 // Keyboard n and N
	s2k["KEY_O"] = HID_KEY_O //0x12 // Keyboard o and O
	s2k["KEY_P"] = HID_KEY_P //0x13 // Keyboard p and P
	s2k["KEY_Q"] = HID_KEY_Q //0x14 // Keyboard q and Q
	s2k["KEY_R"] = HID_KEY_R //0x15 // Keyboard r and R
	s2k["KEY_S"] = HID_KEY_S //0x16 // Keyboard s and S
	s2k["KEY_T"] = HID_KEY_T //0x17 // Keyboard t and T
	s2k["KEY_U"] = HID_KEY_U //0x18 // Keyboard u and U
	s2k["KEY_V"] = HID_KEY_V //0x19 // Keyboard v and V
	s2k["KEY_W"] = HID_KEY_W //0x1a // Keyboard w and W
	s2k["KEY_X"] = HID_KEY_X //0x1b // Keyboard x and X
	s2k["KEY_Y"] = HID_KEY_Y //0x1c // Keyboard y and Y
	s2k["KEY_Z"] = HID_KEY_Z //0x1d // Keyboard z and Z

	s2k["KEY_1"] = HID_KEY_1 //0x1e // Keyboard 1 and !
	s2k["KEY_2"] = HID_KEY_2 //0x1f // Keyboard 2 and @
	s2k["KEY_3"] = HID_KEY_3 //0x20 // Keyboard 3 and #
	s2k["KEY_4"] = HID_KEY_4 //0x21 // Keyboard 4 and $
	s2k["KEY_5"] = HID_KEY_5 //0x22 // Keyboard 5 and %
	s2k["KEY_6"] = HID_KEY_6 //0x23 // Keyboard 6 and ^
	s2k["KEY_7"] = HID_KEY_7 //0x24 // Keyboard 7 and &
	s2k["KEY_8"] = HID_KEY_8 //0x25 // Keyboard 8 and *
	s2k["KEY_9"] = HID_KEY_9 //0x26 // Keyboard 9 and (
	s2k["KEY_0"] = HID_KEY_0 //0x27 // Keyboard 0 and )

	s2k["KEY_ENTER"] = HID_KEY_ENTER           //0x28 // Keyboard Return (ENTER)
	s2k["KEY_ESC"] = HID_KEY_ESC               //0x29 // Keyboard ESCAPE
	s2k["KEY_BACKSPACE"] = HID_KEY_BACKSPACE   //0x2a // Keyboard DELETE (Backspace)
	s2k["KEY_TAB"] = HID_KEY_TAB               //0x2b // Keyboard Tab
	s2k["KEY_SPACE"] = HID_KEY_SPACE           //0x2c // Keyboard Spacebar
	s2k["KEY_MINUS"] = HID_KEY_MINUS           //0x2d // Keyboard - and _
	s2k["KEY_EQUAL"] = HID_KEY_EQUAL           //0x2e // Keyboard = and +
	s2k["KEY_LEFTBRACE"] = HID_KEY_LEFTBRACE   //0x2f // Keyboard [ and {
	s2k["KEY_RIGHTBRACE"] = HID_KEY_RIGHTBRACE //0x30 // Keyboard "] and }
	s2k["KEY_BACKSLASH"] = HID_KEY_BACKSLASH   //0x31 // Keyboard \ and |
	s2k["KEY_HASHTILDE"] = HID_KEY_HASHTILDE   //0x32 // Keyboard Non-US # and ~
	s2k["KEY_SEMICOLON"] = HID_KEY_SEMICOLON   //0x33 // Keyboard ; and :
	s2k["KEY_APOSTROPHE"] = HID_KEY_APOSTROPHE //0x34 // Keyboard ' and "
	s2k["KEY_GRAVE"] = HID_KEY_GRAVE           //0x35 // Keyboard ` and ~
	s2k["KEY_COMMA"] = HID_KEY_COMMA           //0x36 // Keyboard , and <
	s2k["KEY_DOT"] = HID_KEY_DOT               //0x37 // Keyboard . and >
	s2k["KEY_SLASH"] = HID_KEY_SLASH           //0x38 // Keyboard / and ?
	s2k["KEY_CAPSLOCK"] = HID_KEY_CAPSLOCK     //0x39 // Keyboard Caps Lock

	s2k["KEY_F1"] = HID_KEY_F1   //0x3a // Keyboard F1
	s2k["KEY_F2"] = HID_KEY_F2   //0x3b // Keyboard F2
	s2k["KEY_F3"] = HID_KEY_F3   //0x3c // Keyboard F3
	s2k["KEY_F4"] = HID_KEY_F4   //0x3d // Keyboard F4
	s2k["KEY_F5"] = HID_KEY_F5   //0x3e // Keyboard F5
	s2k["KEY_F6"] = HID_KEY_F6   //0x3f // Keyboard F6
	s2k["KEY_F7"] = HID_KEY_F7   //0x40 // Keyboard F7
	s2k["KEY_F8"] = HID_KEY_F8   //0x41 // Keyboard F8
	s2k["KEY_F9"] = HID_KEY_F9   //0x42 // Keyboard F9
	s2k["KEY_F10"] = HID_KEY_F10 //0x43 // Keyboard F10
	s2k["KEY_F11"] = HID_KEY_F11 //0x44 // Keyboard F11
	s2k["KEY_F12"] = HID_KEY_F12 //0x45 // Keyboard F12

	s2k["KEY_SYSRQ"] = HID_KEY_SYSRQ           //0x46 // Keyboard Print Screen
	s2k["KEY_SCROLLLOCK"] = HID_KEY_SCROLLLOCK //0x47 // Keyboard Scroll Lock
	s2k["KEY_PAUSE"] = HID_KEY_PAUSE           //0x48 // Keyboard Pause
	s2k["KEY_INSERT"] = HID_KEY_INSERT         //0x49 // Keyboard Insert
	s2k["KEY_HOME"] = HID_KEY_HOME             //0x4a // Keyboard Home
	s2k["KEY_PAGEUP"] = HID_KEY_PAGEUP         //0x4b // Keyboard Page Up
	s2k["KEY_DELETE"] = HID_KEY_DELETE         //0x4c // Keyboard Delete Forward
	s2k["KEY_END"] = HID_KEY_END               //0x4d // Keyboard End
	s2k["KEY_PAGEDOWN"] = HID_KEY_PAGEDOWN     //0x4e // Keyboard Page Down
	s2k["KEY_RIGHT"] = HID_KEY_RIGHT           //0x4f // Keyboard Right Arrow
	s2k["KEY_LEFT"] = HID_KEY_LEFT             //0x50 // Keyboard Left Arrow
	s2k["KEY_DOWN"] = HID_KEY_DOWN             //0x51 // Keyboard Down Arrow
	s2k["KEY_UP"] = HID_KEY_UP                 //0x52 // Keyboard Up Arrow

	s2k["KEY_NUMLOCK"] = HID_KEY_NUMLOCK       //0x53 // Keyboard Num Lock and Clear
	s2k["KEY_KPSLASH"] = HID_KEY_KPSLASH       //0x54 // Keypad /
	s2k["KEY_KPASTERISK"] = HID_KEY_KPASTERISK //0x55 // Keypad *
	s2k["KEY_KPMINUS"] = HID_KEY_KPMINUS       //0x56 // Keypad -
	s2k["KEY_KPPLUS"] = HID_KEY_KPPLUS         //0x57 // Keypad +
	s2k["KEY_KPENTER"] = HID_KEY_KPENTER       //0x58 // Keypad ENTER
	s2k["KEY_KP1"] = HID_KEY_KP1               //0x59 // Keypad 1 and End
	s2k["KEY_KP2"] = HID_KEY_KP2               //0x5a // Keypad 2 and Down Arrow
	s2k["KEY_KP3"] = HID_KEY_KP3               //0x5b // Keypad 3 and PageDn
	s2k["KEY_KP4"] = HID_KEY_KP4               //0x5c // Keypad 4 and Left Arrow
	s2k["KEY_KP5"] = HID_KEY_KP5               //0x5d // Keypad 5
	s2k["KEY_KP6"] = HID_KEY_KP6               //0x5e // Keypad 6 and Right Arrow
	s2k["KEY_KP7"] = HID_KEY_KP7               //0x5f // Keypad 7 and Home
	s2k["KEY_KP8"] = HID_KEY_KP8               //0x60 // Keypad 8 and Up Arrow
	s2k["KEY_KP9"] = HID_KEY_KP9               //0x61 // Keypad 9 and Page Up
	s2k["KEY_KP0"] = HID_KEY_KP0               //0x62 // Keypad 0 and Insert
	s2k["KEY_KPDOT"] = HID_KEY_KPDOT           //0x63 // Keypad . and Delete

	s2k["KEY_102ND"] = HID_KEY_102ND     //0x64 // Keyboard Non-US \ and |
	s2k["KEY_COMPOSE"] = HID_KEY_COMPOSE //0x65 // Keyboard Application
	s2k["KEY_POWER"] = HID_KEY_POWER     //0x66 // Keyboard Power
	s2k["KEY_KPEQUAL"] = HID_KEY_KPEQUAL //0x67 // Keypad =

	s2k["KEY_F13"] = HID_KEY_F13 //0x68 // Keyboard F13
	s2k["KEY_F14"] = HID_KEY_F14 //0x69 // Keyboard F14
	s2k["KEY_F15"] = HID_KEY_F15 //0x6a // Keyboard F15
	s2k["KEY_F16"] = HID_KEY_F16 //0x6b // Keyboard F16
	s2k["KEY_F17"] = HID_KEY_F17 //0x6c // Keyboard F17
	s2k["KEY_F18"] = HID_KEY_F18 //0x6d // Keyboard F18
	s2k["KEY_F19"] = HID_KEY_F19 //0x6e // Keyboard F19
	s2k["KEY_F20"] = HID_KEY_F20 //0x6f // Keyboard F20
	s2k["KEY_F21"] = HID_KEY_F21 //0x70 // Keyboard F21
	s2k["KEY_F22"] = HID_KEY_F22 //0x71 // Keyboard F22
	s2k["KEY_F23"] = HID_KEY_F23 //0x72 // Keyboard F23
	s2k["KEY_F24"] = HID_KEY_F24 //0x73 // Keyboard F24

	s2k["KEY_OPEN"] = HID_KEY_OPEN             //0x74 // Keyboard Execute
	s2k["KEY_HELP"] = HID_KEY_HELP             //0x75 // Keyboard Help
	s2k["KEY_PROPS"] = HID_KEY_PROPS           //0x76 // Keyboard Menu
	s2k["KEY_FRONT"] = HID_KEY_FRONT           //0x77 // Keyboard Select
	s2k["KEY_STOP"] = HID_KEY_STOP             //0x78 // Keyboard Stop
	s2k["KEY_AGAIN"] = HID_KEY_AGAIN           //0x79 // Keyboard Again
	s2k["KEY_UNDO"] = HID_KEY_UNDO             //0x7a // Keyboard Undo
	s2k["KEY_CUT"] = HID_KEY_CUT               //0x7b // Keyboard Cut
	s2k["KEY_COPY"] = HID_KEY_COPY             //0x7c // Keyboard Copy
	s2k["KEY_PASTE"] = HID_KEY_PASTE           //0x7d // Keyboard Paste
	s2k["KEY_FIND"] = HID_KEY_FIND             //0x7e // Keyboard Find
	s2k["KEY_MUTE"] = HID_KEY_MUTE             //0x7f // Keyboard Mute
	s2k["KEY_VOLUMEUP"] = HID_KEY_VOLUMEUP     //0x80 // Keyboard Volume Up
	s2k["KEY_VOLUMEDOWN"] = HID_KEY_VOLUMEDOWN //0x81 // Keyboard Volume Down
	// = 0x82  Keyboard Locking Caps Lock
	// = 0x83  Keyboard Locking Num Lock
	// = 0x84  Keyboard Locking Scroll Lock
	s2k["KEY_KPCOMMA"] = HID_KEY_KPCOMMA //0x85 // Keypad Comma
	// = 0x86  Keypad Equal Sign
	s2k["KEY_RO"] = HID_KEY_RO                             //0x87 // Keyboard International1
	s2k["KEY_KATAKANAHIRAGANA"] = HID_KEY_KATAKANAHIRAGANA //0x88 // Keyboard International2
	s2k["KEY_YEN"] = HID_KEY_YEN                           //0x89 // Keyboard International3
	s2k["KEY_HENKAN"] = HID_KEY_HENKAN                     //0x8a // Keyboard International4
	s2k["KEY_MUHENKAN"] = HID_KEY_MUHENKAN                 //0x8b // Keyboard International5
	s2k["KEY_KPJPCOMMA"] = HID_KEY_KPJPCOMMA               //0x8c // Keyboard International6
	// = 0x8d  Keyboard International7
	// = 0x8e  Keyboard International8
	// = 0x8f  Keyboard International9
	s2k["KEY_HANGEUL"] = HID_KEY_HANGEUL               //0x90 // Keyboard LANG1
	s2k["KEY_HANJA"] = HID_KEY_HANJA                   //0x91 // Keyboard LANG2
	s2k["KEY_KATAKANA"] = HID_KEY_KATAKANA             //0x92 // Keyboard LANG3
	s2k["KEY_HIRAGANA"] = HID_KEY_HIRAGANA             //0x93 // Keyboard LANG4
	s2k["KEY_ZENKAKUHANKAKU"] = HID_KEY_ZENKAKUHANKAKU //0x94 // Keyboard LANG5
	// = 0x95  Keyboard LANG6
	// = 0x96  Keyboard LANG7
	// = 0x97  Keyboard LANG8
	// = 0x98  Keyboard LANG9
	// = 0x99  Keyboard Alternate Erase
	// = 0x9a  Keyboard SysReq/Attention
	// = 0x9b  Keyboard Cancel
	// = 0x9c  Keyboard Clear
	// = 0x9d  Keyboard Prior
	// = 0x9e  Keyboard Return
	// = 0x9f  Keyboard Separator
	// = 0xa0  Keyboard Out
	// = 0xa1  Keyboard Oper
	// = 0xa2  Keyboard Clear/Again
	// = 0xa3  Keyboard CrSel/Props
	// = 0xa4  Keyboard ExSel

	// = 0xb0  Keypad 00
	// = 0xb1  Keypad 000
	// = 0xb2  Thousands Separator
	// = 0xb3  Decimal Separator
	// = 0xb4  Currency Unit
	// = 0xb5  Currency Sub-unit
	s2k["KEY_KPLEFTPAREN"] = HID_KEY_KPLEFTPAREN   //0xb6 // Keypad (
	s2k["KEY_KPRIGHTPAREN"] = HID_KEY_KPRIGHTPAREN //0xb7 // Keypad )
	// = 0xb8  Keypad {
	// = 0xb9  Keypad }
	// = 0xba  Keypad Tab
	// = 0xbb  Keypad Backspace
	// = 0xbc  Keypad A
	// = 0xbd  Keypad B
	// = 0xbe  Keypad C
	// = 0xbf  Keypad D
	// = 0xc0  Keypad E
	// = 0xc1  Keypad F
	// = 0xc2  Keypad XOR
	// = 0xc3  Keypad ^
	// = 0xc4  Keypad %
	// = 0xc5  Keypad <
	// = 0xc6  Keypad >
	// = 0xc7  Keypad &
	// = 0xc8  Keypad &&
	// = 0xc9  Keypad |
	// = 0xca  Keypad ||
	// = 0xcb  Keypad :
	// = 0xcc  Keypad #
	// = 0xcd  Keypad Space
	// = 0xce  Keypad @
	// = 0xcf  Keypad !
	// = 0xd0  Keypad Memory Store
	// = 0xd1  Keypad Memory Recall
	// = 0xd2  Keypad Memory Clear
	// = 0xd3  Keypad Memory Add
	// = 0xd4  Keypad Memory Subtract
	// = 0xd5  Keypad Memory Multiply
	// = 0xd6  Keypad Memory Divide
	// = 0xd7  Keypad +/-
	// = 0xd8  Keypad Clear
	// = 0xd9  Keypad Clear Entry
	// = 0xda  Keypad Binary
	// = 0xdb  Keypad Octal
	// = 0xdc  Keypad Decimal
	// = 0xdd  Keypad Hexadecimal

	s2k["KEY_LEFTCTRL"] = HID_KEY_LEFTCTRL     //0xe0 // Keyboard Left Control
	s2k["KEY_LEFTSHIFT"] = HID_KEY_LEFTSHIFT   //0xe1 // Keyboard Left Shift
	s2k["KEY_LEFTALT"] = HID_KEY_LEFTALT       //0xe2 // Keyboard Left Alt
	s2k["KEY_LEFTMETA"] = HID_KEY_LEFTMETA     //0xe3 // Keyboard Left GUI
	s2k["KEY_RIGHTCTRL"] = HID_KEY_RIGHTCTRL   //0xe4 // Keyboard Right Control
	s2k["KEY_RIGHTSHIFT"] = HID_KEY_RIGHTSHIFT //0xe5 // Keyboard Right Shift
	s2k["KEY_RIGHTALT"] = HID_KEY_RIGHTALT     //0xe6 // Keyboard Right Alt
	s2k["KEY_RIGHTMETA"] = HID_KEY_RIGHTMETA   //0xe7 // Keyboard Right GUI

	s2k["KEY_MEDIA_PLAYPAUSE"] = HID_KEY_MEDIA_PLAYPAUSE       //0xe8
	s2k["KEY_MEDIA_STOPCD"] = HID_KEY_MEDIA_STOPCD             //0xe9
	s2k["KEY_MEDIA_PREVIOUSSONG"] = HID_KEY_MEDIA_PREVIOUSSONG //0xea
	s2k["KEY_MEDIA_NEXTSONG"] = HID_KEY_MEDIA_NEXTSONG         //0xeb
	s2k["KEY_MEDIA_EJECTCD"] = HID_KEY_MEDIA_EJECTCD           //0xec
	s2k["KEY_MEDIA_VOLUMEUP"] = HID_KEY_MEDIA_VOLUMEUP         //0xed
	s2k["KEY_MEDIA_VOLUMEDOWN"] = HID_KEY_MEDIA_VOLUMEDOWN     //0xee
	s2k["KEY_MEDIA_MUTE"] = HID_KEY_MEDIA_MUTE                 //0xef
	s2k["KEY_MEDIA_WWW"] = HID_KEY_MEDIA_WWW                   //0xf0
	s2k["KEY_MEDIA_BACK"] = HID_KEY_MEDIA_BACK                 //0xf1
	s2k["KEY_MEDIA_FORWARD"] = HID_KEY_MEDIA_FORWARD           //0xf2
	s2k["KEY_MEDIA_STOP"] = HID_KEY_MEDIA_STOP                 //0xf3
	s2k["KEY_MEDIA_FIND"] = HID_KEY_MEDIA_FIND                 //0xf4
	s2k["KEY_MEDIA_SCROLLUP"] = HID_KEY_MEDIA_SCROLLUP         //0xf5
	s2k["KEY_MEDIA_SCROLLDOWN"] = HID_KEY_MEDIA_SCROLLDOWN     //0xf6
	s2k["KEY_MEDIA_EDIT"] = HID_KEY_MEDIA_EDIT                 //0xf7
	s2k["KEY_MEDIA_SLEEP"] = HID_KEY_MEDIA_SLEEP               //0xf8
	s2k["KEY_MEDIA_COFFEE"] = HID_KEY_MEDIA_COFFEE             //0xf9
	s2k["KEY_MEDIA_REFRESH"] = HID_KEY_MEDIA_REFRESH           //0xfa
	s2k["KEY_MEDIA_CALC"] = HID_KEY_MEDIA_CALC                 //0xfb

	return
}

func NaiveAsciiTransform(mod HIDMod, key HIDKey) (res string) {
	//ASCII
	if key >= HID_KEY_A && key <= HID_KEY_Z {
		if mod&(HID_MOD_KEY_RIGHT_SHIFT|HID_MOD_KEY_LEFT_SHIFT) > 0 {
			return string((byte(key) + (0x41 - 0x04)))
		} else {
			return string((byte(key) + (0x61 - 0x04)))
		}
	}

	if mod&(HID_MOD_KEY_RIGHT_SHIFT|HID_MOD_KEY_LEFT_SHIFT) == 0 {
		if key >= HID_KEY_1 && key <= HID_KEY_9 {
			return string((byte(key) + (0x31 - 0x1e)))
		}

		if key == HID_KEY_0 {
			return string(byte(0x30))
		}
	}

	if key == HID_KEY_ENTER {
		return " \n"
	}
	if key == HID_KEY_TAB {
		return " \t"
	}
	if key == HID_KEY_SPACE {
		return " "
	}

	return ""
}

func NaiveKeymodTransform(in uint8) (mod HIDMod, key HIDKey) {
	//ASCII
	chr := in
	if chr >= 0x61 && chr <= 0x61+26 {
		return 0x00, HIDKey(chr - 0x61 + 0x04)
	}
	if chr >= 0x41 && chr <= 0x41+26 {
		return HID_MOD_KEY_LEFT_SHIFT, HIDKey(chr - 0x41 + 0x04)
	}
	if chr >= 0x31 && chr <= 0x40 {
		return 0x00, HIDKey(chr - 0x31 + 0x1e)
	}
	if chr == 0x30 {
		return 0x00, HID_KEY_0
	}
	if in == '\n' {
		return 0x00, HID_KEY_ENTER
	}
	if in == '\t' {
		return 0x00, HID_KEY_TAB
	}
	if in == ' ' {
		return 0x00, HID_KEY_SPACE
	}

	return
}
