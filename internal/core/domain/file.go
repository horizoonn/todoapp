package domain

type File struct {
	buffer []byte
}

func (f *File) Buffer() []byte {
	return f.buffer
}

func NewFile(buffer []byte) File {
	return File{
		buffer: buffer,
	}
}
