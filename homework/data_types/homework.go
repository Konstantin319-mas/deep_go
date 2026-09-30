package main

// endianTypes — ограничение на поддерживаемые целочисленные типы.
type endianTypes interface {
	~uint16 | ~uint32 | ~uint64
}

// ToLittleEndian конвертирует число из Big Endian (прямого порядка
// следования байт) в Little Endian (обратный порядок следования байт).
func ToLittleEndian[T endianTypes](n T) T {
	switch v := any(n).(type) {
	case uint16:
		return T(swap16(v))
	case uint32:
		return T(swap32(v))
	case uint64:
		return T(swap64(v))
	default:
		panic("ToLittleEndian: unsupported type")
	}
}

// swap16 меняет байты uint16 местами: 0xABCD -> 0xCDAB.
func swap16(n uint16) uint16 {
	return (n << 8) | (n >> 8)
}

// swap32 меняет порядок байт uint32: 0x01020304 -> 0x04030201.
// 1) меняются местами соседние пары байт (b0<->b1, b2<->b3),
// 2) меняются местами 16-битные половины (b0<->b2, b1<->b3).
func swap32(n uint32) uint32 {
	x := ((n & 0xFF00FF00) >> 8) | ((n & 0x00FF00FF) << 8)
	return (x >> 16) | (x << 16)
}

// swap64 меняет порядок байт uint64.
func swap64(n uint64) uint64 {
	x := ((n & 0x00FF00FF00FF00FF) << 8) | ((n & 0xFF00FF00FF00FF00) >> 8)
	x = ((x & 0x0000FFFF0000FFFF) << 16) | ((x & 0xFFFF0000FFFF0000) >> 16)
	return (x << 32) | (x >> 32)
}
