package ipc

type request string

const (
	StatusRequest       request = "status"
	ListAccountsRequest request = "list"
	QRCodeImportRequest request = "qrcodeimport"
	TextImportRequest   request = "textimport"
	CopyCodeRequest     request = "copycode"
)
