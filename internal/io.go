package internal

type BytesCounter struct {
	count int64
}

func (counter *BytesCounter) Write(p []byte) (int, error) {
	counter.count += int64(len(p))
	return len(p), nil
}

func (counter *BytesCounter) Count() int64 {
	return counter.count
}
