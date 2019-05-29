package frame

// StartSentinel The sentinel to indicate start of data transmission
const StartSentinel uint32 = 0xDEADBEEF

// Header for the data transmission, holding fields applicable for this logical 'frame'
type Header struct {
	NumLEDs uint16
}

// LEDInfo The data portion of the frame
type LEDInfo struct {
	Row        uint8
	Column     uint8
	Red        uint8
	Green      uint8
	Blue       uint8
	Brightness uint8
}
