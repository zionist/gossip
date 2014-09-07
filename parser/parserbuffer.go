package parser

import "bytes"

const c_writeBuffSize int = 5

type parserBuffer struct {
	buffer bytes.Buffer

	lineBreaks []int

	dataIn       chan string
	requestsIn   chan dataRequest
	requestQueue []dataRequest

	stop chan bool
}

func newParserBuffer() *parserBuffer {
	var pb parserBuffer
	pb.lineBreaks = make([]int, 0)
	pb.requestsIn = make(chan dataRequest, 0)
	pb.requestQueue = make([]dataRequest, 0)

	go pb.manage()

	return &pb
}

func (pb *parserBuffer) NextLine() string {
	var request lineRequest = make(chan string)
	pb.requestsIn <- request
	return <-request
}

func (pb *parserBuffer) manage() {
	for {
		for len(pb.requestQueue) > 0 {
			// See if we can respond to any requests.
			switch pb.requestQueue[0].(type) {
			case lineRequest:
				if len(pb.lineBreaks) > 0 {
					breakpoint := pb.lineBreaks[0]
					pb.requestQueue[0].(lineRequest) <- string(pb.buffer.Next(breakpoint))
					pb.lineBreaks = pb.lineBreaks[1:]
					for idx := range pb.lineBreaks {
						pb.lineBreaks[idx] -= breakpoint
					}
				} else {
					// Don't service any subsequent requests, as we can't yet process the first
					// one, and they need to be handled in order. Wait for more data and try again.
					break
				}
			case dataRequest:
				// TODO
			}
		}

		// No pending requests to process, so just wait for new data or requests.
		select {
		case data := <-pb.dataIn:
			bufferEndIdx := pb.buffer.Len()
			pb.buffer.WriteString(data)
			for idx := range indexAll(data, "\r\n") {
				pb.lineBreaks = append(pb.lineBreaks, bufferEndIdx+idx)
			}
		case request := <-pb.requestsIn:
			pb.requestQueue = append(pb.requestQueue, request)
		case <-pb.stop:
			// Stop main loop, dispatch all pending requests, and end.
			break
		}
	}

	// TODO Dispose
}

type dataRequest interface{}

type lineRequest chan string

type chunkRequest struct {
	n        int
	response chan string
}

func indexAll(source string, target string) []int {
	return nil // TODO
}
