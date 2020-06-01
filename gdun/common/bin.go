package common

func HexStringToBytes(s string) []byte {
	n := len(s)
	byteLength := n / 2

	r := make([]byte, byteLength)

	i := 0
	j := 0
	b := (byte)(0)
	for (i < n) && (j < byteLength) {
		switch s[i] {
		case '1':
			b = (byte)(1)
		case '2':
			b = (byte)(2)
		case '3':
			b = (byte)(3)
		case '4':
			b = (byte)(4)
		case '5':
			b = (byte)(5)
		case '6':
			b = (byte)(6)
		case '7':
			b = (byte)(7)
		case '8':
			b = (byte)(8)
		case '9':
			b = (byte)(9)
		case 'a', 'A':
			b = (byte)(10)
		case 'b', 'B':
			b = (byte)(11)
		case 'c', 'C':
			b = (byte)(12)
		case 'd', 'D':
			b = (byte)(13)
		case 'e', 'E':
			b = (byte)(14)
		case 'f', 'F':
			b = (byte)(15)
		default:
			b = (byte)(0)
		}
		i++

		if i%2 == 1 {
			r[j] = b
		} else {
			r[j] = r[j] << 4
			r[j] = r[j] | b
			j++
		}
	}

	return r
}

func BytesToInt32(b []byte, start int) int {
	return int(b[start]<<24 + b[start+1]<<16 + b[start+2]<<8 + b[start+3])
}
