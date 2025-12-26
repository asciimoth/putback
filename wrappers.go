package putback

import "io"

type PutBackReader struct {
	io.Reader
	Buffer BackBuffer
}

func (pb *PutBackReader) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReader) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.Reader, p)
}

type PutBackReadCloser struct {
	io.ReadCloser
	Buffer BackBuffer
}

func (pb *PutBackReadCloser) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReadCloser) Close() error {
	pb.Buffer.Wipe()
	return pb.ReadCloser.Close()
}

func (pb *PutBackReadCloser) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadCloser, p)
}

type PutBackReadWriter struct {
	io.ReadWriter
	Buffer BackBuffer
}

func (pb *PutBackReadWriter) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReadWriter) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadWriter, p)
}

type PutBackReadWriteCloser struct {
	io.ReadWriteCloser
	Buffer BackBuffer
}

func (pb *PutBackReadWriteCloser) PutBack(bytes []byte) {
	pb.Buffer.PutBack(bytes)
}

func (pb *PutBackReadWriteCloser) Close() error {
	pb.Buffer.Wipe()
	return pb.ReadWriteCloser.Close()
}

func (pb *PutBackReadWriteCloser) Read(p []byte) (n int, err error) {
	return readJoin(&pb.Buffer, pb.ReadWriteCloser, p)
}
