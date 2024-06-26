package pkzip

type Compression uint16

const (
	Stored Compression = iota
	Shrunk
	ReducedFactor1
	ReducedFactor2
	ReducedFactor3
	ReducedFactor4
	Imploded
	Reserved
	Deflated
	EnhancedDeflated
	PKWareDataCompressionLibraryImplode
	BZIP2
	Reserved2
	LZMA
	Reserved3
	Reserved4
	IBMTERSE
	IBMLZ77z
)

const (
	PPMd1 Compression = iota + 98
)

func (c Compression) String() string {
	switch c {
	case Stored:
		return "Stored"
	case Shrunk:
		return "Shrunk"
	case ReducedFactor1:
		return "Reduced with factor 1"
	case ReducedFactor2:
		return "Reduced with factor 2"
	case ReducedFactor3:
		return "Reduced with factor 3"
	case ReducedFactor4:
		return "Reduced with factor 4"
	case Imploded:
		return "Imploded"
	case Reserved:
		return "Reserved"
	case Deflated:
		return "Deflated"
	case EnhancedDeflated:
		return "Enhanced Deflated"
	case PKWareDataCompressionLibraryImplode:
		return "PKWare Data Compression Library Imploded"
	case BZIP2:
		return "BZIP2"
	case Reserved2:
		return "Reserved"
	case LZMA:
		return "LZMA"
	case Reserved3:
		return "Reserved"
	case Reserved4:
		return "Reserved"
	case IBMTERSE:
		return "IBM TERSE"
	case IBMLZ77z:
		return "IBM LZ77z"
	case PPMd1:
		return "PPMd version I, Rev 1"
	default:
		return "Unknown"
	}
}
