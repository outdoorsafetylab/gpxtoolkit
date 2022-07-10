package xml

type Stack struct {
	slice []string
}

func (s *Stack) Contains(a *Stack) bool {
	if len(s.slice) < len(a.slice) {
		return false
	}
	for i, e := range a.slice {
		if s.slice[i] != e {
			return false
		}
	}
	return true
}

func (s *Stack) Clone() *Stack {
	c := &Stack{slice: make([]string, len(s.slice))}
	copy(c.slice, s.slice)
	return c
}

func (s *Stack) Depth() int {
	return len(s.slice)
}

func (s *Stack) Push(e string) {
	s.slice = append(s.slice, e)
}

func (s *Stack) Peek() string {
	n := len(s.slice)
	if n <= 0 {
		return ""
	} else {
		return s.slice[n-1]
	}
}

func (s *Stack) Pop() string {
	i := len(s.slice) - 1
	last := s.slice[i]
	s.slice = s.slice[:i]
	return last
}
