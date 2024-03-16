package sds

type Sds []byte

func Empty() Sds {
	return Sds{}
}

func (s Sds) Cat(t []byte) Sds {
	return append(s, t...)
}

func (s Sds) Catstr(ss string) Sds {
	return s.Cat([]byte(ss))
}

func (s Sds) Len() int {
	return len(s)
}

func (s Sds) Isempty() bool {
	return len(s) > 0
}
