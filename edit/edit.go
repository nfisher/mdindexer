package edit

// Distance2 calculates the levenshtein distance using a 2 row matrix.
func Distance2(s1, s2 string) int {
	m := len(s1) + 1
	n := len(s2) + 1
	mat := newMatrix(2, n)
	row := mat[0]
	for i := 1; i < m; i++ {
		ch := s1[i-1]
		prevRow := row
		row = mat[i%2]
		row[0] = i
		for j := 1; j < n; j++ {
			min := prevRow[j-1]
			if ch != s2[j-1] {
				min += 2
			}
			candidate := prevRow[j] + 1
			if candidate < min {
				min = candidate
			}
			candidate = row[j-1] + 1
			if candidate < min {
				min = candidate
			}
			row[j] = min
		}
	}
	return mat[(m-1)%2][n-1]
}

// Distance calculates the levenshtein distance using a full matrix.
func Distance(s1, s2 string) int {
	m := len(s1) + 1
	n := len(s2) + 1
	mat := newMatrix(m, n)
	for i := 1; i < m; i++ {
		ch := s1[i-1]
		row := mat[i]
		prevRow := mat[i-1]
		for j := 1; j < n; j++ {
			min := prevRow[j-1]
			if ch != s2[j-1] {
				min += 2
			}
			candidate := prevRow[j] + 1
			if candidate < min {
				min = candidate
			}
			candidate = row[j-1] + 1
			if candidate < min {
				min = candidate
			}
			row[j] = min
		}
	}
	return mat[m-1][n-1]
}

func newMatrix(m, n int) [][]int {
	mat := make([][]int, m)
	for i := range mat {
		mat[i] = make([]int, n)
		mat[i][0] = i
		if i > 0 {
			continue
		}
		for j := range mat[i] {
			mat[i][j] = j
		}
	}
	return mat
}
