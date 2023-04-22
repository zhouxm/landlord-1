package common

type MySortInt [][]int

func (p MySortInt) Len() int {
	return len(p)
}

func (p MySortInt) Less(i, j int) bool {
	lenPi := len(p[i])
	lenPj := len(p[j])
	if lenPi > lenPj {
		return true
	} else if lenPi == lenPj {
		for k := 0; k < lenPi; k++ {
			if p[i][k] < p[j][k] {
				return true
			} else if p[i][k] == p[j][k] {
				continue
			} else {
				return false
			}
		}
	} else {
		return false
	}
	return true
}

func (p MySortInt) Swap(i, j int) {
	p[i], p[j] = p[j], p[i]
}
