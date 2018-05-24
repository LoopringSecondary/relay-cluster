package utils

import "container/list"

//https://medium.com/capital-one-developers/building-an-unbounded-channel-in-go-789e175cd2cd
func MakeInfinite() (chan<- interface{}, <-chan interface{}) {
	in := make(chan interface{})
	out := make(chan interface{})
	go func() {
		buffer := list.New()
		outCh := func() chan<- interface{} {
			if buffer.Len() == 0 {
				return nil
			}
			return out
		}
		curVal := func() interface{} {
			if buffer.Len() == 0 {
				return nil
			}
			return buffer.Front().Value
		}
		for buffer.Len() > 0 || in != nil {
			select {
			case v, ok := <-in:
				if !ok {
					in = nil
				} else {
					buffer.PushBack(v)
				}

			case outCh() <- curVal(): //will block if buffer is empty
				head := buffer.Front()
				buffer.Remove(head)
			}
		}
	}()
	return in, out
}
