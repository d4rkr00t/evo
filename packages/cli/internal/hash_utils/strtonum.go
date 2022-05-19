package hash_utils

func StrToNum(s string) int {
	var hash = 0

	for pos, char := range s {
		hash += (pos*prime(char) + int(char-'0'))
	}

	return hash
}

func prime(char rune) int {
	var primes = []int{13, 17, 19, 23, 29, 31, 37, 41, 43, 47, 53, 59, 61, 67, 71, 73, 79, 83, 97, 101, 103, 107, 109, 113, 127, 131}
	var charCode = int(char - '0')
	var idx = pmod(charCode, len(primes))
	return primes[idx]
}

func pmod(x, d int) int {
	x = x % d
	if x >= 0 {
		return x
	}
	if d < 0 {
		return x - d
	}
	return x + d
}
