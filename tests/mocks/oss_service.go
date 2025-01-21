package mocks

import "mime/multipart"

type MockOSSService struct {
	UploadAvatarFunc func(file *multipart.FileHeader) (string, error)
}

func (m *MockOSSService) UploadAvatar(file *multipart.FileHeader) (string, error) {
	return m.UploadAvatarFunc(file)
}
