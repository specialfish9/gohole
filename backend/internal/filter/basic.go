package filter

type BasicFilter map[string]struct{}

var _ Filter = (BasicFilter)(nil)

func NewBasic(domains []string) Filter {
	f := make(BasicFilter)
	for _, d := range domains {
		f[d] = struct{}{}
	}

	return f
}

func (f BasicFilter) Filter(q string) (bool, error) {
	_, ok := f[q]
	return ok, nil
}

func (f BasicFilter) Size() int {
	return len(f)
}
