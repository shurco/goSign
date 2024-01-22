package sign

import (
	"strconv"
	"strings"
)

func (context *SignContext) createInfo() (info string, err error) {
	var builder strings.Builder

	original_info := context.PDFReader.Trailer().Key("Info")
	builder.WriteString(strconv.Itoa(int(context.InfoData.ObjectId)) + " 0 obj\n<<")

	info_keys := original_info.Keys()
	for _, key := range info_keys {
		builder.WriteString("/" + key)
		if key == "ModDate" {
			builder.WriteString(pdfDateTime(context.SignData.Signature.Info.Date))
		} else {
			builder.WriteString(pdfString(original_info.Key(key).RawString()))
		}
	}

	builder.WriteString(">>\nendobj\n")
	return builder.String(), nil
}
